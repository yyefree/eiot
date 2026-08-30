import 'dart:convert';
import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';

class DeviceDetailPage extends StatefulWidget {
  final int deviceId;
  const DeviceDetailPage({super.key, required this.deviceId});

  @override
  State<DeviceDetailPage> createState() => _DeviceDetailPageState();
}

class _DeviceDetailPageState extends State<DeviceDetailPage> with SingleTickerProviderStateMixin {
  DeviceDetail? _detail;
  bool _loading = true;
  final Map<String, dynamic> _controlValues = {};
  final Map<String, TextEditingController> _textControllers = {};

  late TabController _tabController;

  List<dynamic> _events = [];
  bool _eventsLoading = false;
  int _eventsPage = 1;
  bool _hasMoreEvents = true;

  List<dynamic> _serviceHistory = [];
  bool _serviceHistoryLoading = false;

  Map<String, dynamic>? _shadow;
  bool _shadowLoading = false;

  List<dynamic> _dataHistory = [];
  bool _dataHistoryLoading = false;

  final Map<String, TextEditingController> _serviceInputControllers = {};

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 5, vsync: this);
    _load();
  }

  @override
  void dispose() {
    for (final c in _textControllers.values) {
      c.dispose();
    }
    for (final c in _serviceInputControllers.values) {
      c.dispose();
    }
    _textControllers.clear();
    _serviceInputControllers.clear();
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    if (!mounted) return;
    setState(() => _loading = true);
    final resp = await ApiClient.get('/device/${widget.deviceId}');
    if (!mounted) return;
    if (resp.ok && resp.data != null) {
      final data = resp.data as Map<String, dynamic>;
      final deviceData = data['device'];
      if (deviceData == null || deviceData is! Map<String, dynamic>) {
        setState(() => _loading = false);
        return;
      }
      final device = DeviceItem.fromJson(deviceData);
      final latest = (data['latest'] as Map<String, dynamic>?)?.map((k, v) => MapEntry(k, v?.toString() ?? '')) ?? {};
      final tm = data['thingModel'] as Map<String, dynamic>?;
      final propsRaw = tm?['properties'] as List? ?? data['properties'] as List? ?? [];
      final props = propsRaw
          .whereType<Map<String, dynamic>>()
          .map((e) => ThingProperty.fromJson(e))
          .toList();
      final eventsRaw = tm?['events'] as List? ?? [];
      final events = eventsRaw.whereType<Map<String, dynamic>>().map((e) => ThingEvent.fromJson(e)).toList();
      final servicesRaw = tm?['services'] as List? ?? [];
      final services = servicesRaw.whereType<Map<String, dynamic>>().map((e) => ThingService.fromJson(e)).toList();
      _detail = DeviceDetail(device: device, latest: latest, properties: props, events: events, services: services);
      for (final p in props) {
        if (p.canWrite) {
          if (p.isBool) {
            _controlValues[p.identifier] = latest[p.identifier] == 'true';
          } else if (p.isNumber) {
            _controlValues[p.identifier] = double.tryParse(latest[p.identifier] ?? '') ?? p.min ?? 0;
          } else {
            _controlValues[p.identifier] = latest[p.identifier] ?? '';
          }
        }
      }
    }
    if (mounted) setState(() => _loading = false);
  }

  Future<void> _loadEvents({bool append = false}) async {
    if (_eventsLoading) return;
    setState(() => _eventsLoading = true);
    final resp = await ApiClient.getDeviceEvents(widget.deviceId, page: _eventsPage, size: 20);
    if (!mounted) return;
    if (resp.ok && resp.data != null) {
      final d = resp.data as Map<String, dynamic>;
      final list = d['list'] as List? ?? d['data'] as List? ?? [];
      setState(() {
        if (append) {
          _events.addAll(list);
        } else {
          _events = list;
        }
        _hasMoreEvents = list.length >= 20;
      });
    }
    setState(() => _eventsLoading = false);
  }

  Future<void> _loadServiceHistory() async {
    setState(() => _serviceHistoryLoading = true);
    final resp = await ApiClient.getDeviceServices(widget.deviceId, page: 1, size: 50);
    if (!mounted) return;
    if (resp.ok && resp.data != null) {
      final d = resp.data as Map<String, dynamic>;
      setState(() {
        _serviceHistory = d['list'] as List? ?? d['data'] as List? ?? [];
      });
    }
    setState(() => _serviceHistoryLoading = false);
  }

  Future<void> _loadShadow() async {
    setState(() => _shadowLoading = true);
    final resp = await ApiClient.getDeviceShadow(widget.deviceId);
    if (!mounted) return;
    if (resp.ok && resp.data != null) {
      setState(() {
        _shadow = resp.data as Map<String, dynamic>;
      });
    }
    setState(() => _shadowLoading = false);
  }

  Future<void> _loadDataHistory() async {
    setState(() => _dataHistoryLoading = true);
    final resp = await ApiClient.getDeviceDataHistory(_detail!.device.deviceSn, limit: 100);
    if (!mounted) return;
    if (resp.ok && resp.data != null) {
      final d = resp.data as Map<String, dynamic>;
      setState(() {
        _dataHistory = d['list'] as List? ?? d['data'] as List? ?? [];
      });
    }
    setState(() => _dataHistoryLoading = false);
  }

  Future<void> _sendCommand(ThingProperty p) async {
    final resp = await ApiClient.post('/device/${widget.deviceId}/control', {
      p.identifier: _controlValues[p.identifier],
    });
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(resp.ok ? '${p.name} 指令已下发' : '下发失败: ${resp.msg}')),
      );
      if (resp.ok) _load();
    }
  }

  Future<void> _simulateEvent(ThingEvent evt) async {
    final output = <String, dynamic>{};
    for (final param in evt.outputParams) {
      if (param.dataType == 'int') {
        output[param.identifier] = 0;
      } else if (param.dataType == 'float' || param.dataType == 'double') {
        output[param.identifier] = 0.0;
      } else if (param.dataType == 'bool') {
        output[param.identifier] = false;
      } else {
        output[param.identifier] = '';
      }
    }
    final resp = await ApiClient.reportDeviceEvent(widget.deviceId, evt.identifier, evt.name, output);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(resp.ok ? '事件已上报' : '上报失败: ${resp.msg}')),
      );
    }
  }

  Future<void> _invokeService(ThingService svc) async {
    final input = <String, dynamic>{};
    for (final param in svc.inputParams) {
      final key = '${svc.identifier}_${param.identifier}';
      final val = _serviceInputControllers[key]?.text ?? '';
      if (param.dataType == 'int') {
        input[param.identifier] = int.tryParse(val) ?? 0;
      } else if (param.dataType == 'float' || param.dataType == 'double') {
        input[param.identifier] = double.tryParse(val) ?? 0.0;
      } else if (param.dataType == 'bool') {
        input[param.identifier] = val.toLowerCase() == 'true';
      } else {
        input[param.identifier] = val;
      }
    }
    final resp = await ApiClient.invokeDeviceService(widget.deviceId, svc.identifier, svc.name, input);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(resp.ok ? '服务调用成功' : '调用失败: ${resp.msg}')),
      );
      if (resp.ok) _loadServiceHistory();
    }
  }

  TextEditingController _getController(ThingProperty p) {
    return _textControllers.putIfAbsent(p.identifier, () {
      final c = TextEditingController(text: '${_controlValues[p.identifier] ?? ''}');
      return c;
    });
  }

  TextEditingController _getServiceInputController(String key, {String initial = ''}) {
    return _serviceInputControllers.putIfAbsent(key, () => TextEditingController(text: initial));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('设备详情'),
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: const [
            Tab(text: '属性'),
            Tab(text: '事件'),
            Tab(text: '服务'),
            Tab(text: '历史'),
            Tab(text: '影子'),
          ],
        ),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _detail == null
              ? const Center(child: Text('加载失败'))
              : TabBarView(
                  controller: _tabController,
                  children: [
                    _buildPropertiesTab(),
                    _buildEventsTab(),
                    _buildServicesTab(),
                    _buildHistoryTab(),
                    _buildShadowTab(),
                  ],
                ),
    );
  }

  Widget _buildPropertiesTab() {
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(_detail!.device.online ? Icons.wifi : Icons.wifi_off,
                          color: _detail!.device.online ? Colors.green : Colors.grey),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(_detail!.device.name,
                            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                      ),
                      Chip(
                        label: Text(_detail!.device.online ? '在线' : '离线'),
                        backgroundColor: _detail!.device.online ? Colors.green.shade50 : Colors.grey.shade100,
                      ),
                    ],
                  ),
                  const Divider(height: 24),
                  _infoRow('产品', _detail!.device.productName),
                  _infoRow('DeviceSN', _detail!.device.deviceSn),
                  _infoRow('ProductKey', _detail!.device.productKey),
                  _infoRow('创建时间', _detail!.device.createdAt),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          if (_detail!.latest.isNotEmpty) ...[
            const Text('实时数据', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Card(
              child: Column(
                children: _detail!.latest.entries
                    .map((e) => ListTile(
                          dense: true,
                          title: Text(e.key),
                          trailing: Text(e.value, style: const TextStyle(fontWeight: FontWeight.bold)),
                        ))
                    .toList(),
              ),
            ),
            const SizedBox(height: 16),
          ],
          if (_detail!.properties.isNotEmpty) ...[
            const Text('物模型控制', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            ..._detail!.properties.where((p) => p.canWrite).map((p) => Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text('${p.name} (${p.unit})',
                                  style: const TextStyle(fontWeight: FontWeight.bold)),
                            ),
                            if (p.required)
                              const Text('*', style: TextStyle(color: Colors.red, fontSize: 16)),
                          ],
                        ),
                        if (p.description.isNotEmpty)
                          Padding(
                            padding: const EdgeInsets.only(top: 4),
                            child: Text(p.description, style: TextStyle(color: Colors.grey.shade600, fontSize: 12)),
                          ),
                        const SizedBox(height: 8),
                        if (p.isBool)
                          Switch(
                            value: _controlValues[p.identifier] == true,
                            onChanged: (v) {
                              setState(() => _controlValues[p.identifier] = v);
                            },
                          )
                        else if (p.isEnum && p.enumValues != null)
                          DropdownButtonFormField<String>(
                            value: _controlValues[p.identifier]?.toString(),
                            items: p.enumValues!
                                .map((v) => DropdownMenuItem(value: v, child: Text(v)))
                                .toList(),
                            onChanged: (v) {
                              setState(() => _controlValues[p.identifier] = v);
                            },
                            decoration: const InputDecoration(border: OutlineInputBorder()),
                          )
                        else if (p.isNumber)
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Slider(
                                value: (_controlValues[p.identifier] as num?)?.toDouble() ?? (p.min ?? 0),
                                min: p.min ?? 0,
                                max: p.max ?? 100,
                                divisions: ((p.max ?? 100) - (p.min ?? 0)) > 0
                                    ? ((p.max ?? 100) - (p.min ?? 0)).round()
                                    : 100,
                                label: '${_controlValues[p.identifier]}',
                                onChanged: (v) {
                                  setState(() => _controlValues[p.identifier] = v);
                                },
                              ),
                              Text('值: ${_controlValues[p.identifier]}  范围: ${p.min ?? 0} ~ ${p.max ?? 0}',
                                  style: const TextStyle(fontSize: 12)),
                            ],
                          )
                        else
                          TextField(
                            controller: _getController(p),
                            onChanged: (v) => _controlValues[p.identifier] = v,
                            decoration: InputDecoration(
                              border: const OutlineInputBorder(),
                              hintText: p.name,
                            ),
                          ),
                        const SizedBox(height: 8),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton(
                            onPressed: () => _sendCommand(p),
                            child: const Text('下发指令'),
                          ),
                        ),
                      ],
                    ),
                  ),
                )),
          ],
        ],
      ),
    );
  }

  Widget _buildEventsTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              const Expanded(
                child: Text('事件定义', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              ),
              ElevatedButton.icon(
                icon: const Icon(Icons.refresh, size: 16),
                label: const Text('刷新'),
                onPressed: () => _loadEvents(),
              ),
            ],
          ),
        ),
        if (_detail!.events.isNotEmpty)
          SizedBox(
            height: 160,
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              itemCount: _detail!.events.length,
              itemBuilder: (context, index) {
                final evt = _detail!.events[index];
                return Card(
                  margin: const EdgeInsets.only(right: 8),
                  child: SizedBox(
                    width: 220,
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(evt.name, style: const TextStyle(fontWeight: FontWeight.bold)),
                              ),
                              _eventTypeBadge(evt.type),
                            ],
                          ),
                          const SizedBox(height: 4),
                          Text(evt.identifier, style: const TextStyle(fontSize: 12, color: Colors.grey)),
                          if (evt.description.isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 4),
                              child: Text(evt.description, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                            ),
                          const Spacer(),
                          SizedBox(
                            width: double.infinity,
                            child: OutlinedButton.icon(
                              icon: const Icon(Icons.send, size: 14),
                              label: const Text('模拟上报', style: TextStyle(fontSize: 12)),
                              onPressed: () => _simulateEvent(evt),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          )
        else
          const Padding(
            padding: EdgeInsets.all(24),
            child: Text('无事件定义'),
          ),
        const Divider(),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Row(
            children: [
              const Expanded(
                child: Text('事件记录', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              ),
              TextButton(
                onPressed: () {
                  _eventsPage = 1;
                  _loadEvents();
                },
                child: const Text('刷新记录'),
              ),
            ],
          ),
        ),
        Expanded(
          child: _eventsLoading && _events.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : _events.isEmpty
                  ? const Center(child: Text('暂无事件记录'))
                  : ListView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      itemCount: _events.length,
                      itemBuilder: (context, index) {
                        final e = _events[index];
                        return Card(
                          margin: const EdgeInsets.only(bottom: 8),
                          child: ListTile(
                            title: Text(e['event_name'] ?? e['event_id'] ?? '未知事件'),
                            subtitle: Text(
                              '时间: ${e['created_at'] ?? e['timestamp'] ?? '-'}',
                              style: const TextStyle(fontSize: 12),
                            ),
                            trailing: _eventTypeBadge(e['type'] ?? 'info'),
                          ),
                        );
                      },
                    ),
        ),
      ],
    );
  }

  Widget _buildServicesTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              const Expanded(
                child: Text('服务列表', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              ),
              TextButton.icon(
                icon: const Icon(Icons.history, size: 16),
                label: const Text('调用历史'),
                onPressed: () {
                  _loadServiceHistory();
                  _showServiceHistorySheet();
                },
              ),
            ],
          ),
        ),
        Expanded(
          child: _detail!.services.isEmpty
              ? const Center(child: Text('无服务定义'))
              : ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  itemCount: _detail!.services.length,
                  itemBuilder: (context, index) {
                    final svc = _detail!.services[index];
                    return Card(
                      margin: const EdgeInsets.only(bottom: 12),
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Expanded(
                                  child: Text(svc.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                                ),
                                Chip(
                                  label: Text(svc.callType == 'sync' ? '同步' : '异步', style: const TextStyle(fontSize: 11)),
                                  materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                ),
                              ],
                            ),
                            if (svc.description.isNotEmpty)
                              Padding(
                                padding: const EdgeInsets.only(top: 4),
                                child: Text(svc.description, style: TextStyle(color: Colors.grey.shade600, fontSize: 12)),
                              ),
                            Text(svc.identifier, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                            const SizedBox(height: 12),
                            if (svc.inputParams.isNotEmpty) ...[
                              const Text('输入参数:', style: TextStyle(fontWeight: FontWeight.w500, fontSize: 13)),
                              const SizedBox(height: 8),
                              ...svc.inputParams.map((param) {
                                final key = '${svc.identifier}_${param.identifier}';
                                return Padding(
                                  padding: const EdgeInsets.only(bottom: 8),
                                  child: TextField(
                                    controller: _getServiceInputController(key),
                                    decoration: InputDecoration(
                                      labelText: '${param.name} (${param.dataType}${param.unit.isNotEmpty ? ', ${param.unit}' : ''})',
                                      border: const OutlineInputBorder(),
                                      isDense: true,
                                    ),
                                  ),
                                );
                              }),
                            ],
                            const SizedBox(height: 8),
                            SizedBox(
                              width: double.infinity,
                              child: ElevatedButton.icon(
                                icon: const Icon(Icons.play_arrow, size: 18),
                                label: const Text('调用'),
                                onPressed: () => _invokeService(svc),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  Widget _buildHistoryTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              const Expanded(
                child: Text('历史数据', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              ),
              ElevatedButton.icon(
                icon: const Icon(Icons.refresh, size: 16),
                label: const Text('刷新'),
                onPressed: () => _loadDataHistory(),
              ),
            ],
          ),
        ),
        Expanded(
          child: _dataHistoryLoading
              ? const Center(child: CircularProgressIndicator())
              : _dataHistory.isEmpty
                  ? const Center(child: Text('暂无历史数据，点击刷新加载'))
                  : ListView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      itemCount: _dataHistory.length,
                      itemBuilder: (context, index) {
                        final item = _dataHistory[index];
                        return Card(
                          margin: const EdgeInsets.only(bottom: 6),
                          child: ListTile(
                            dense: true,
                            title: Text(item['property'] ?? item['key'] ?? '-'),
                            subtitle: Text(item['created_at'] ?? item['timestamp'] ?? '-', style: const TextStyle(fontSize: 11)),
                            trailing: Text(
                              item['value']?.toString() ?? '-',
                              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                            ),
                          ),
                        );
                      },
                    ),
        ),
      ],
    );
  }

  Widget _buildShadowTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              const Expanded(
                child: Text('设备影子', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              ),
              ElevatedButton.icon(
                icon: const Icon(Icons.refresh, size: 16),
                label: const Text('刷新'),
                onPressed: () => _loadShadow(),
              ),
            ],
          ),
        ),
        Expanded(
          child: _shadowLoading
              ? const Center(child: CircularProgressIndicator())
              : _shadow == null
                  ? const Center(child: Text('暂无影子数据，点击刷新加载'))
                  : SingleChildScrollView(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          if (_shadow!['version'] != null)
                            Padding(
                              padding: const EdgeInsets.only(bottom: 12),
                              child: Chip(label: Text('版本: ${_shadow!['version']}')),
                            ),
                          _shadowSection('期望值 (desired)', _shadow!['desired']),
                          const SizedBox(height: 12),
                          _shadowSection('上报值 (reported)', _shadow!['reported']),
                        ],
                      ),
                    ),
        ),
      ],
    );
  }

  Widget _shadowSection(String title, dynamic data) {
    String jsonStr;
    if (data is Map || data is List) {
      const encoder = JsonEncoder.withIndent('  ');
      jsonStr = encoder.convert(data);
    } else if (data != null) {
      jsonStr = data.toString();
    } else {
      jsonStr = '{}';
    }
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 8),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.grey.shade200),
              ),
              child: SelectableText(
                jsonStr,
                style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showServiceHistorySheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) {
        return DraggableScrollableSheet(
          initialChildSize: 0.6,
          minChildSize: 0.3,
          maxChildSize: 0.9,
          expand: false,
          builder: (context, scrollController) {
            return Column(
              children: [
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      const Expanded(
                        child: Text('服务调用历史', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close),
                        onPressed: () => Navigator.pop(context),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: _serviceHistoryLoading
                      ? const Center(child: CircularProgressIndicator())
                      : _serviceHistory.isEmpty
                          ? const Center(child: Text('暂无调用历史'))
                          : ListView.builder(
                              controller: scrollController,
                              padding: const EdgeInsets.symmetric(horizontal: 16),
                              itemCount: _serviceHistory.length,
                              itemBuilder: (context, index) {
                                final item = _serviceHistory[index];
                                return Card(
                                  margin: const EdgeInsets.only(bottom: 8),
                                  child: ListTile(
                                    title: Text(item['service_name'] ?? item['service_id'] ?? '-'),
                                    subtitle: Text(
                                      '时间: ${item['created_at'] ?? item['timestamp'] ?? '-'}',
                                      style: const TextStyle(fontSize: 12),
                                    ),
                                    trailing: Icon(
                                      item['status'] == 'success' ? Icons.check_circle : Icons.error,
                                      color: item['status'] == 'success' ? Colors.green : Colors.red,
                                      size: 20,
                                    ),
                                  ),
                                );
                              },
                            ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  Widget _eventTypeBadge(String type) {
    Color color;
    switch (type) {
      case 'warning':
        color = Colors.orange;
        break;
      case 'error':
        color = Colors.red;
        break;
      default:
        color = Colors.blue;
    }
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Text(type, style: TextStyle(color: color, fontSize: 10, fontWeight: FontWeight.w500)),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(width: 80, child: Text(label, style: const TextStyle(color: Colors.grey))),
          Expanded(child: Text(value.isEmpty ? '-' : value)),
        ],
      ),
    );
  }
}
