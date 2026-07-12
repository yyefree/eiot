import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/auth_provider.dart';
import 'pages/login_page.dart';
import 'pages/device_list_page.dart';
import 'pages/share_page.dart';
import 'pages/profile_page.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const EiotApp());
}

class EiotApp extends StatelessWidget {
  const EiotApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => AuthProvider()..init(),
      child: MaterialApp(
        title: 'EIOT 物联网',
        theme: ThemeData(
          colorSchemeSeed: const Color(0xFF409EFF),
          useMaterial3: true,
        ),
        home: const _Root(),
        routes: {
          '/login': (_) => const LoginPage(),
          '/home': (_) => const _MainShell(),
        },
      ),
    );
  }
}

class _Root extends StatefulWidget {
  const _Root();

  @override
  State<_Root> createState() => _RootState();
}

class _RootState extends State<_Root> {
  bool _navigated = false;

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    // 等待初始化完成
    if (!auth.initialized) {
      return const Scaffold(
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.devices_other, size: 72, color: Color(0xFF409EFF)),
              SizedBox(height: 16),
              Text('EIOT 物联网平台', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
              SizedBox(height: 32),
              CircularProgressIndicator(),
            ],
          ),
        ),
      );
    }
    // 根据登录状态导航
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_navigated) {
        _navigated = true;
        Navigator.pushReplacementNamed(
          context,
          auth.isLoggedIn ? '/home' : '/login',
        );
      }
    });
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}

class _MainShell extends StatefulWidget {
  const _MainShell();

  @override
  State<_MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<_MainShell> {
  int _index = 0;

  final _pages = const [
    DeviceListPage(),
    SharePage(),
    ProfilePage(),
  ];

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    if (!auth.isLoggedIn) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        Navigator.pushReplacementNamed(context, '/login');
      });
    }
    return Scaffold(
      body: _pages[_index],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: const [
          NavigationDestination(icon: Icon(Icons.devices), label: '设备'),
          NavigationDestination(icon: Icon(Icons.share), label: '共享'),
          NavigationDestination(icon: Icon(Icons.person), label: '我的'),
        ],
      ),
    );
  }
}
