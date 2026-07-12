import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final user = auth.user;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        const SizedBox(height: 32),
        CircleAvatar(
          radius: 48,
          child: Text(
            (user?.nickname ?? '?')[0].toUpperCase(),
            style: const TextStyle(fontSize: 36),
          ),
        ),
        const SizedBox(height: 16),
        Center(
          child: Text(
            user?.nickname ?? '未登录',
            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
          ),
        ),
        const SizedBox(height: 8),
        Center(
          child: Text(user?.phone ?? '', style: const TextStyle(color: Colors.grey)),
        ),
        const SizedBox(height: 32),
        Card(
          child: Column(
            children: [
              ListTile(
                leading: const Icon(Icons.person, color: Colors.blue),
                title: const Text('用户名'),
                trailing: Text(user?.username ?? '-'),
              ),
              ListTile(
                leading: const Icon(Icons.phone, color: Colors.green),
                title: const Text('手机号'),
                trailing: Text(user?.phone ?? '-'),
              ),
              ListTile(
                leading: const Icon(Icons.badge, color: Colors.orange),
                title: const Text('角色'),
                trailing: Chip(
                  label: Text(user?.isAdmin == true ? '管理员' : '普通用户'),
                  backgroundColor: user?.isAdmin == true ? Colors.orange.shade50 : Colors.blue.shade50,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        ElevatedButton.icon(
          onPressed: () async {
            await auth.logout();
          },
          icon: const Icon(Icons.logout),
          label: const Text('退出登录'),
          style: ElevatedButton.styleFrom(
            foregroundColor: Colors.red,
            minimumSize: const Size(double.infinity, 48),
          ),
        ),
      ],
    );
  }
}
