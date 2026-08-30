import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/auth_provider.dart';
import 'providers/home_provider.dart';
import 'providers/scene_provider.dart';
import 'providers/message_provider.dart';
import 'pages/login_page.dart';
import 'pages/home_page.dart';
import 'pages/scene_page.dart';
import 'pages/add_device_page.dart';
import 'pages/messages_page.dart';
import 'pages/profile_page.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const EiotApp());
}

class EiotApp extends StatelessWidget {
  const EiotApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthProvider()..init()),
        ChangeNotifierProvider(create: (_) => HomeProvider()),
        ChangeNotifierProvider(create: (_) => SceneProvider()),
        ChangeNotifierProvider(create: (_) => MessageProvider()),
      ],
      child: MaterialApp(
        title: '云智能',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          colorSchemeSeed: const Color(0xFF007DFF),
          useMaterial3: true,
          scaffoldBackgroundColor: const Color(0xFFF5F5F5),
          cardTheme: CardTheme(
            color: Colors.white,
            elevation: 0,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
          appBarTheme: const AppBarTheme(
            backgroundColor: Colors.white,
            foregroundColor: Color(0xFF333333),
            elevation: 0,
            centerTitle: true,
            titleTextStyle: TextStyle(
              color: Color(0xFF333333),
              fontSize: 18,
              fontWeight: FontWeight.w600,
            ),
          ),
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
    if (!auth.initialized) {
      return const Scaffold(
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.devices_other, size: 72, color: Color(0xFF007DFF)),
              SizedBox(height: 16),
              Text('云智能', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
              SizedBox(height: 32),
              CircularProgressIndicator(),
            ],
          ),
        ),
      );
    }
    if (!_navigated) {
      _navigated = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        Navigator.pushReplacementNamed(
          context,
          auth.isLoggedIn ? '/home' : '/login',
        );
      });
    }
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
    HomePage(),
    ScenePage(),
    AddDevicePage(),
    MessagesPage(),
    ProfilePage(),
  ];

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final auth = context.read<AuthProvider>();
      if (auth.isLoggedIn) {
        context.read<HomeProvider>().loadHomes();
        context.read<SceneProvider>().loadScenes();
        context.read<MessageProvider>().loadMessages();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    if (!auth.isLoggedIn) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        Navigator.pushReplacementNamed(context, '/login');
      });
    }
    return Scaffold(
      body: IndexedStack(index: _index, children: _pages),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) {
          if (i == 2) {
            Navigator.push(context, MaterialPageRoute(builder: (_) => const AddDevicePage()));
          } else {
            setState(() => _index = i);
          }
        },
        backgroundColor: Colors.white,
        elevation: 8,
        indicatorColor: const Color(0xFF007DFF).withOpacity(0.1),
        height: 60,
        labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined, color: Color(0xFF999999)),
            selectedIcon: Icon(Icons.home, color: Color(0xFF007DFF)),
            label: '首页',
          ),
          NavigationDestination(
            icon: Icon(Icons.auto_awesome_outlined, color: Color(0xFF999999)),
            selectedIcon: Icon(Icons.auto_awesome, color: Color(0xFF007DFF)),
            label: '智能',
          ),
          NavigationDestination(
            icon: SizedBox.shrink(),
            label: '',
          ),
          NavigationDestination(
            icon: Icon(Icons.chat_bubble_outline, color: Color(0xFF999999)),
            selectedIcon: Icon(Icons.chat_bubble, color: Color(0xFF007DFF)),
            label: '消息',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline, color: Color(0xFF999999)),
            selectedIcon: Icon(Icons.person, color: Color(0xFF007DFF)),
            label: '我的',
          ),
        ],
      ),
      floatingActionButton: SizedBox(
        width: 48,
        height: 48,
        child: FloatingActionButton(
          onPressed: () {
            Navigator.push(context, MaterialPageRoute(builder: (_) => const AddDevicePage()));
          },
          backgroundColor: const Color(0xFF007DFF),
          elevation: 4,
          shape: const CircleBorder(),
          child: const Icon(Icons.add, color: Colors.white, size: 28),
        ),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
    );
  }
}
