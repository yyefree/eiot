import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';

class DeviceDetailPage extends StatefulWidget {
  final int deviceId;
  const DeviceDetailPage({super.key, required this.deviceId});

  @override
  State<DeviceDetailPage> createState() => _DeviceDetailPageState();
}

class _DeviceDetailPageState extends State<DeviceDetailPage> {
  DeviceDetail? _detail;
  bool _loading = true;
  final Map<String, dynamic> _controlValues = {};

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    final resp = await ApiClient.get('/device/${widget.deviceId}');
    if (resp.ok && resp.data != null) {
      final data = resp.data as Map<String, dynamic>;
      final device = DeviceItem.fromJson(data['device'] as Map<String, dynamic>);
      final latest = (data['latest'] as Map<String, dynamic>?)?.map((k, v) => MapEntry(k, v.toString())) ?? {};
      final tm = data['thingModel'] as Map<String, dynamic>?;
      final propsRaw = tm?['properties'] as List? ?? data['properties'] as List? ?? [];
      final props = propsRaw.map((e) => ThingProperty.fromJson(e as Map<String, dynamic>)).toList();
      _detail = DeviceDetail(device: device, latest: latest, properties: props);
      // 初始化控制值
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
    setState(() => _loading = false);
  }

  Future<void> _sendCommand(ThingProperty p) async {
    final resp = await ApiClient.post('/device/${widget.deviceId}/control', {
      'params': {p.identifier: _controlValues[p.identifier]},
    });
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(resp.ok ? '${p.name} 指令已下发' : '下发失败: ${resp.msg}')),
      );
      if (resp.ok) _load();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('设备详情')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _detail == null
              ? const Center(child: Text('加载失败'))
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      // 基本信息
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
                                    backgroundColor: _detail!.device.online
                                        ? Colors.green.shade50
                                        : Colors.grey.shade100,
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
                      // 最新数据
                      if (_detail!.latest.isNotEmpty) ...[
                        const Text('实时数据', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                        const SizedBox(height: 8),
                        Card(
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              children: _detail!.latest.entries.map((e) {
                                final prop = _detail!.properties.where((p) => p.identifier == e.key).firstOrNull;
                                return Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 6),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                    children: [
                                      Text(prop?.name ?? e.key),
                                      Text('${e.value} ${prop?.unit ?? ''}',
                                          style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.blue)),
                                    ],
                                  ),
                                );
                              }).toList(),
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],
                      // 控制面板
                      ..._buildControls(),
                    ],
                  ),
                ),
    );
  }

  List<Widget> _buildControls() {
    final writable = _detail!.properties.where((p) => p.canWrite).toList();
    if (writable.isEmpty) return [];
    return [
      const Text('设备控制', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
      const SizedBox(height: 8),
      Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: writable.map((p) {
              return Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: _buildControl(p),
              );
            }).toList(),
          ),
        ),
      ),
    ];
  }

  Widget _buildControl(ThingProperty p) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('${p.name}${p.unit.isNotEmpty ? ' (${p.unit})' : ''}',
            style: const TextStyle(fontWeight: FontWeight.w500)),
        const SizedBox(height: 8),
        if (p.isBool)
          Row(
            children: [
              Switch(
                value: _controlValues[p.identifier] ?? false,
                onChanged: (v) => setState(() => _controlValues[p.identifier] = v),
              ),
              const Spacer(),
              ElevatedButton(onPressed: () => _sendCommand(p), child: const Text('下发')),
            ],
          )
        else if (p.isNumber)
          Row(
            children: [
              Expanded(
                child: Slider(
                  value: (_controlValues[p.identifier] as num?)?.toDouble() ?? p.min ?? 0,
                  min: p.min ?? 0,
                  max: p.max ?? 100,
                  divisions: ((p.max ?? 100) - (p.min ?? 0)).round(),
                  label: _controlValues[p.identifier].toString(),
                  onChanged: (v) => setState(() => _controlValues[p.identifier] = v),
                ),
              ),
              SizedBox(
                width: 60,
                child: Text('${_controlValues[p.identifier]?.toStringAsFixed(1)}',
                    style: const TextStyle(fontWeight: FontWeight.bold)),
              ),
              ElevatedButton(onPressed: () => _sendCommand(p), child: const Text('下发')),
            ],
          )
        else
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: TextEditingController(text: _controlValues[p.identifier]?.toString() ?? ''),
                  onChanged: (v) => _controlValues[p.identifier] = v,
                  decoration: const InputDecoration(border: OutlineInputBorder(), isDense: true),
                ),
              ),
              const SizedBox(width: 8),
              ElevatedButton(onPressed: () => _sendCommand(p), child: const Text('下发')),
            ],
          ),
      ],
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(width: 100, child: Text(label, style: const TextStyle(color: Colors.grey))),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}
