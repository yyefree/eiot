import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';
import 'device_detail_page.dart';

class DeviceListPage extends StatefulWidget {
  const DeviceListPage({super.key});

  @override
  State<DeviceListPage> createState() => _DeviceListPageState();
}

class _DeviceListPageState extends State<DeviceListPage> {
  List<DeviceItem> _devices = [];
  bool _loading = true;
  int _page = 1;
  int _total = 0;
  final _scrollCtrl = ScrollController();

  @override
  void initState() {
    super.initState();
    _load();
    _scrollCtrl.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollCtrl.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollCtrl.position.pixels >= _scrollCtrl.position.maxScrollExtent - 100 &&
        !_loading && _devices.length < _total) {
      _page++;
      _load(more: true);
    }
  }

  Future<void> _load({bool more = false}) async {
    if (!more) {
      _page = 1;
      setState(() => _loading = true);
    }
    final resp = await ApiClient.get('/device', params: {'page': _page, 'size': 20});
    if (resp.ok && resp.data != null) {
      final data = resp.data as Map<String, dynamic>;
      final list = (data['list'] as List?)?.map((e) => DeviceItem.fromJson(e as Map<String, dynamic>)).toList() ?? [];
      _total = data['total'] as int? ?? 0;
      setState(() {
        if (more) {
          _devices.addAll(list);
        } else {
          _devices = list;
        }
        _loading = false;
      });
    } else {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return RefreshIndicator(
      onRefresh: _load,
      child: _loading && _devices.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : _devices.isEmpty
              ? ListView(
                  children: const [
                    SizedBox(height: 200),
                    Center(child: Text('暂无设备', style: TextStyle(color: Colors.grey, fontSize: 16))),
                  ],
                )
              : ListView.builder(
                  controller: _scrollCtrl,
                  padding: const EdgeInsets.all(12),
                  itemCount: _devices.length + (_devices.length < _total ? 1 : 0),
                  itemBuilder: (ctx, i) {
                    if (i >= _devices.length) {
                      return const Padding(
                        padding: EdgeInsets.all(16),
                        child: Center(child: CircularProgressIndicator()),
                      );
                    }
                    final d = _devices[i];
                    return Card(
                      child: ListTile(
                        leading: Icon(
                          d.online ? Icons.wifi : Icons.wifi_off,
                          color: d.online ? Colors.green : Colors.grey,
                        ),
                        title: Text(d.name),
                        subtitle: Text(d.productName, style: const TextStyle(fontSize: 13)),
                        trailing: Chip(
                          label: Text(d.online ? '在线' : '离线', style: const TextStyle(fontSize: 12)),
                          backgroundColor: d.online ? Colors.green.shade50 : Colors.grey.shade100,
                        ),
                        onTap: () async {
                          await Navigator.push(ctx,
                              MaterialPageRoute(builder: (_) => DeviceDetailPage(deviceId: d.id)));
                          _load();
                        },
                      ),
                    );
                  },
                ),
    );
  }
}
