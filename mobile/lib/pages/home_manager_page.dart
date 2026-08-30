import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/home_provider.dart';
import '../models/models.dart';
import 'room_manager_page.dart';

class HomeManagerPage extends StatefulWidget {
  const HomeManagerPage({super.key});

  @override
  State<HomeManagerPage> createState() => _HomeManagerPageState();
}

class _HomeManagerPageState extends State<HomeManagerPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<HomeProvider>().loadHomes();
    });
  }

  @override
  Widget build(BuildContext context) {
    final homeProvider = context.watch<HomeProvider>();

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(title: const Text('家庭管理')),
      body: homeProvider.loading
          ? const Center(child: CircularProgressIndicator())
          : homeProvider.homes.isEmpty
              ? _buildEmptyState()
              : RefreshIndicator(
                  onRefresh: homeProvider.loadHomes,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(12),
                    itemCount: homeProvider.homes.length,
                    itemBuilder: (context, index) => _buildHomeCard(homeProvider.homes[index], homeProvider),
                  ),
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCreateHomeDialog(homeProvider),
        backgroundColor: const Color(0xFF007DFF),
        child: const Icon(Icons.add, color: Colors.white),
      ),
    );
  }

  Widget _buildHomeCard(Home home, HomeProvider provider) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () {
          provider.switchHome(home);
          Navigator.push(context, MaterialPageRoute(
            builder: (_) => RoomManagerPage(home: home),
          ));
        },
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: const Color(0xFF007DFF).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.home, color: Color(0xFF007DFF), size: 24),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(home.name, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
                            if (home.isDefault) ...[
                              const SizedBox(width: 8),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF007DFF).withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: const Text('默认', style: TextStyle(fontSize: 10, color: Color(0xFF007DFF))),
                              ),
                            ],
                          ],
                        ),
                        const SizedBox(height: 4),
                        Text(
                          home.address.isEmpty ? '${home.memberCount} 位成员' : '${home.address} · ${home.memberCount} 位成员',
                          style: const TextStyle(fontSize: 12, color: Color(0xFF999999)),
                        ),
                      ],
                    ),
                  ),
                  PopupMenuButton<String>(
                    icon: const Icon(Icons.more_vert, color: Color(0xFF999999)),
                    onSelected: (value) {
                      if (value == 'edit') _showEditHomeDialog(home, provider);
                      if (value == 'delete') _confirmDeleteHome(home, provider);
                    },
                    itemBuilder: (context) => [
                      const PopupMenuItem(value: 'edit', child: Text('编辑')),
                      const PopupMenuItem(value: 'delete', child: Text('删除', style: TextStyle(color: Colors.red))),
                    ],
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.home_outlined, size: 80, color: const Color(0xFFDDDDDD)),
          const SizedBox(height: 16),
          const Text('暂无家庭', style: TextStyle(fontSize: 16, color: Color(0xFF999999))),
          const SizedBox(height: 8),
          const Text('点击右下角 + 创建家庭', style: TextStyle(fontSize: 13, color: Color(0xFFCCCCCC))),
        ],
      ),
    );
  }

  void _showCreateHomeDialog(HomeProvider provider) {
    final nameCtrl = TextEditingController();
    final addrCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('创建家庭'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: nameCtrl, decoration: const InputDecoration(labelText: '家庭名称', hintText: '我的家')),
            const SizedBox(height: 12),
            TextField(controller: addrCtrl, decoration: const InputDecoration(labelText: '地址', hintText: '可选')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              if (nameCtrl.text.isNotEmpty) {
                await provider.createHome(nameCtrl.text, addrCtrl.text);
                if (ctx.mounted) Navigator.pop(ctx);
              }
            },
            child: const Text('创建'),
          ),
        ],
      ),
    );
  }

  void _showEditHomeDialog(Home home, HomeProvider provider) {
    final nameCtrl = TextEditingController(text: home.name);
    final addrCtrl = TextEditingController(text: home.address);
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('编辑家庭'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: nameCtrl, decoration: const InputDecoration(labelText: '家庭名称')),
            const SizedBox(height: 12),
            TextField(controller: addrCtrl, decoration: const InputDecoration(labelText: '地址')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              if (nameCtrl.text.isNotEmpty) {
                await provider.updateHome(home.id, nameCtrl.text, addrCtrl.text);
                if (ctx.mounted) Navigator.pop(ctx);
              }
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _confirmDeleteHome(Home home, HomeProvider provider) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除家庭'),
        content: Text('确定要删除"${home.name}"吗？此操作不可恢复。'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              await provider.deleteHome(home.id);
              if (ctx.mounted) Navigator.pop(ctx);
            },
            child: const Text('删除', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }
}
