import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../api/api_client.dart';
import 'home_manager_page.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final user = auth.user;

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      body: ListView(
        padding: EdgeInsets.zero,
        children: [
          Container(
            padding: const EdgeInsets.fromLTRB(20, 60, 20, 24),
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [Color(0xFF007DFF), Color(0xFF0055CC)],
              ),
            ),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 36,
                  backgroundColor: Colors.white.withOpacity(0.2),
                  child: Text(
                    (user?.nickname?.isNotEmpty == true ? user!.nickname! : '?')[0].toUpperCase(),
                    style: const TextStyle(fontSize: 28, color: Colors.white, fontWeight: FontWeight.bold),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        user?.nickname ?? '未登录',
                        style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.white),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        user?.phone ?? '',
                        style: const TextStyle(fontSize: 14, color: Colors.white70),
                      ),
                    ],
                  ),
                ),
                Icon(Icons.chevron_right, color: Colors.white.withOpacity(0.7)),
              ],
            ),
          ),
          const SizedBox(height: 12),
          _buildMenuGroup([
            _menuItem(Icons.home, '家庭管理', () {
              Navigator.push(context, MaterialPageRoute(builder: (_) => const HomeManagerPage()));
            }),
            _divider(),
            _menuItem(Icons.share, '设备共享', () {
              _showShareDialog(context);
            }),
            _divider(),
            _menuItem(Icons.system_update, '固件升级', () {
              _showFirmwareInfo(context);
            }),
          ]),
          const SizedBox(height: 12),
          _buildMenuGroup([
            _menuItem(Icons.lock_outline, '修改密码', () {
              _showChangePasswordDialog(context);
            }),
            _divider(),
            _menuItem(Icons.language, '多语言', () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('多语言功能开发中'), duration: Duration(seconds: 1)),
              );
            }),
            _divider(),
            _menuItem(Icons.info_outline, '关于', () {
              _showAboutDialog(context);
            }),
          ]),
          const SizedBox(height: 24),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: OutlinedButton(
              onPressed: () async {
                final confirmed = await showDialog<bool>(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('退出登录'),
                    content: const Text('确定要退出登录吗？'),
                    actions: [
                      TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('取消')),
                      TextButton(
                        onPressed: () => Navigator.pop(ctx, true),
                        child: const Text('确定', style: TextStyle(color: Colors.red)),
                      ),
                    ],
                  ),
                );
                if (confirmed == true && context.mounted) {
                  await auth.logout();
                }
              },
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.red,
                side: const BorderSide(color: Colors.red),
                minimumSize: const Size(double.infinity, 48),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              child: const Text('退出登录'),
            ),
          ),
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  void _showChangePasswordDialog(BuildContext context) {
    final oldCtrl = TextEditingController();
    final newCtrl = TextEditingController();
    final confirmCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('修改密码'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: oldCtrl, obscureText: true, decoration: const InputDecoration(labelText: '旧密码', border: OutlineInputBorder())),
            const SizedBox(height: 12),
            TextField(controller: newCtrl, obscureText: true, decoration: const InputDecoration(labelText: '新密码', border: OutlineInputBorder())),
            const SizedBox(height: 12),
            TextField(controller: confirmCtrl, obscureText: true, decoration: const InputDecoration(labelText: '确认新密码', border: OutlineInputBorder())),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              if (newCtrl.text != confirmCtrl.text) {
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('两次密码不一致')));
                return;
              }
              if (newCtrl.text.length < 6) {
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('密码长度不能少于6位')));
                return;
              }
              final resp = await ApiClient.put('/user/password', {
                'old_password': oldCtrl.text,
                'new_password': newCtrl.text,
              });
              if (ctx.mounted) Navigator.pop(ctx);
              if (resp.ok) {
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('密码修改成功')));
              } else {
                ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(resp.msg)));
              }
            },
            child: const Text('确认'),
          ),
        ],
      ),
    );
  }

  void _showShareDialog(BuildContext context) {
    final snCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('设备共享'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('输入对方手机号共享您的设备', style: TextStyle(color: Color(0xFF999999), fontSize: 13)),
            const SizedBox(height: 12),
            TextField(controller: snCtrl, decoration: const InputDecoration(labelText: '对方手机号', border: OutlineInputBorder())),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              if (snCtrl.text.isEmpty) return;
              final resp = await ApiClient.get('/device/share', params: {'page': 1, 'size': 50});
              if (ctx.mounted) Navigator.pop(ctx);
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text(resp.ok ? '共享列表已加载' : resp.msg)),
              );
            },
            child: const Text('查看共享'),
          ),
        ],
      ),
    );
  }

  void _showFirmwareInfo(BuildContext context) async {
    final resp = await ApiClient.getOTAFirmwareList();
    if (context.mounted) {
      final list = (resp.data is Map<String, dynamic> ? (resp.data['list'] ?? []) : []) as List;
      showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('固件升级'),
          content: SizedBox(
            width: double.maxFinite,
            child: list.isEmpty
                ? const Text('暂无可用固件', style: TextStyle(color: Color(0xFF999999)))
                : ListView.builder(
                    shrinkWrap: true,
                    itemCount: list.length,
                    itemBuilder: (_, i) {
                      final fw = list[i];
                      return ListTile(
                        title: Text('v${fw['version'] ?? ''}'),
                        subtitle: Text(fw['changelog'] ?? '无更新说明'),
                        trailing: Text(fw['status'] ?? '', style: const TextStyle(fontSize: 12)),
                      );
                    },
                  ),
          ),
          actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('关闭'))],
        ),
      );
    }
  }

  void _showAboutDialog(BuildContext context) {
    showAboutDialog(
      context: context,
      applicationName: '云智能',
      applicationVersion: '1.0.0',
      applicationLegalese: '© 2026 飞燕IoT平台',
      children: [
        const Text('基于阿里云飞燕平台标准开发的物联网管理应用'),
        const SizedBox(height: 8),
        const Text('支持设备管理、场景联动、OTA升级等功能'),
      ],
    );
  }

  Widget _buildMenuGroup(List<Widget> children) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(children: children),
    );
  }

  Widget _menuItem(IconData icon, String title, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            Icon(icon, size: 22, color: const Color(0xFF007DFF)),
            const SizedBox(width: 12),
            Expanded(child: Text(title, style: const TextStyle(fontSize: 15))),
            const Icon(Icons.chevron_right, size: 20, color: Color(0xFFCCCCCC)),
          ],
        ),
      ),
    );
  }

  Widget _divider() {
    return const Divider(height: 1, indent: 50, color: Color(0xFFF0F0F0));
  }
}
