import 'package:flutter/material.dart';
import '../api/api_client.dart';

class SharePage extends StatefulWidget {
  const SharePage({super.key});

  @override
  State<SharePage> createState() => _SharePageState();
}

class _SharePageState extends State<SharePage> {
  List<dynamic> _shares = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    if (!mounted) return;
    setState(() => _loading = true);
    final resp = await ApiClient.get('/device/share', params: {'page': 1, 'size': 50});
    if (!mounted) return;
    if (resp.ok && resp.data != null) {
      final data = resp.data as Map<String, dynamic>;
      _shares = data['list'] as List? ?? [];
    }
    setState(() => _loading = false);
  }

  Future<void> _revoke(int id) async {
    final resp = await ApiClient.delete('/device/share/$id');
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(resp.ok ? '已撤销共享' : '操作失败: ${resp.msg}')),
    );
    if (resp.ok) _load();
  }

  @override
  Widget build(BuildContext context) {
    return RefreshIndicator(
      onRefresh: _load,
      child: _loading
          ? const Center(child: CircularProgressIndicator())
          : _shares.isEmpty
              ? ListView(children: const [
                  SizedBox(height: 200),
                  Center(child: Text('暂无共享记录', style: TextStyle(color: Colors.grey))),
                ])
              : ListView.builder(
                  padding: const EdgeInsets.all(12),
                  itemCount: _shares.length,
                  itemBuilder: (ctx, i) {
                    final s = _shares[i];
                    if (s is! Map<String, dynamic>) return const SizedBox.shrink();
                    final id = s['id'];
                    return Card(
                      child: ListTile(
                        leading: const Icon(Icons.share, color: Colors.blue),
                        title: Text('设备 #${s['device_id'] ?? '-'}'),
                        subtitle: Text('共享给用户 #${s['share_user_id'] ?? '-'}'),
                        trailing: id != null
                            ? IconButton(
                                icon: const Icon(Icons.delete_outline, color: Colors.red),
                                onPressed: () {
                                  final shareId = id is int ? id : int.tryParse('$id');
                                  if (shareId != null) _revoke(shareId);
                                },
                              )
                            : null,
                      ),
                    );
                  },
                ),
    );
  }
}
