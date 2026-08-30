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
  final Map<String, TextEditingController> _textControllers = {};

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    for (final c in _textControllers.values) {
      c.dispose();
    }
    _textControllers.clear();
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
      _detail = DeviceDetail(device: device, latest: latest, properties: props);
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

  TextEditingController _getController(ThingProperty p) {
    return _textControllers.putIfAbsent(p.identifier, () {
      final c = TextEditingController(text: '${_controlValues[p.identifier] ?? ''}');
      return c;
    });
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
                      if (_detail!.latest.isNotEmpty) ...[
                        const Text('实时数据', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                        const SizedBox(height: 8),
                        Card(
                          child: Column(
                            children: _detail!.latest.entries.map((e) => ListTile(
                              dense: true,
                              title: Text(e.key),
                              trailing: Text(e.value, style: const TextStyle(fontWeight: FontWeight.bold)),
                            )).toList(),
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
                                Text('${p.name} (${p.unit})',
                                    style: const TextStyle(fontWeight: FontWeight.bold)),
                                const SizedBox(height: 8),
                                if (p.isBool)
                                  Switch(
                                    value: _controlValues[p.identifier] == true,
                                    onChanged: (v) {
                                      setState(() => _controlValues[p.identifier] = v);
                                    },
                                  )
                                else if (p.isNumber)
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
                ),
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
