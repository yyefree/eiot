import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  int _tab = 0;
  final _phoneCtrl = TextEditingController();
  final _pwdCtrl = TextEditingController();
  final _codeCtrl = TextEditingController();
  int _countdown = 0;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _pwdCtrl.dispose();
    _codeCtrl.dispose();
    super.dispose();
  }

  void _startCountdown() {
    _countdown = 60;
    setState(() {});
    Future.doWhile(() async {
      await Future.delayed(const Duration(seconds: 1));
      _countdown--;
      if (!mounted) return false;
      setState(() {});
      return _countdown > 0;
    });
  }

  Future<void> _sendCode() async {
    if (_phoneCtrl.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请输入手机号')));
      return;
    }
    final auth = context.read<AuthProvider>();
    final code = await auth.sendCode(_phoneCtrl.text);
    if (code != null) {
      _startCountdown();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('验证码: $code (5分钟内有效)')));
      }
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('发送失败')));
      }
    }
  }

  Future<void> _login() async {
    final auth = context.read<AuthProvider>();
    bool ok;
    if (_tab == 0) {
      ok = await auth.login(_phoneCtrl.text, _pwdCtrl.text);
    } else {
      ok = await auth.loginByCode(_phoneCtrl.text, _codeCtrl.text);
    }
    if (ok && mounted) {
      Navigator.pushReplacementNamed(context, '/home');
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('登录失败，请检查账号密码')));
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.devices_other, size: 64, color: Theme.of(context).primaryColor),
              const SizedBox(height: 16),
              const Text('EIOT 物联网平台', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              const Text('智慧物联，尽在掌控', style: TextStyle(color: Colors.grey)),
              const SizedBox(height: 32),
              DefaultTabController(
                length: 2,
                child: TabBar(
                  tabs: const [Tab(text: '密码登录'), Tab(text: '验证码登录')],
                  onTap: (i) => setState(() => _tab = i),
                  labelColor: Theme.of(context).primaryColor,
                  unselectedLabelColor: Colors.grey,
                ),
              ),
              const SizedBox(height: 24),
              TextField(
                controller: _phoneCtrl,
                decoration: const InputDecoration(
                  labelText: '手机号',
                  prefixIcon: Icon(Icons.phone),
                  border: OutlineInputBorder(),
                  hintText: '13800000000',
                ),
                keyboardType: TextInputType.phone,
              ),
              const SizedBox(height: 16),
              if (_tab == 0)
                TextField(
                  controller: _pwdCtrl,
                  decoration: const InputDecoration(
                    labelText: '密码',
                    prefixIcon: Icon(Icons.lock),
                    border: OutlineInputBorder(),
                    hintText: 'admin123',
                  ),
                  obscureText: true,
                )
              else
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _codeCtrl,
                        decoration: const InputDecoration(
                          labelText: '验证码',
                          prefixIcon: Icon(Icons.sms),
                          border: OutlineInputBorder(),
                        ),
                        keyboardType: TextInputType.number,
                      ),
                    ),
                    const SizedBox(width: 12),
                    ElevatedButton(
                      onPressed: _countdown > 0 ? null : _sendCode,
                      child: Text(_countdown > 0 ? '${_countdown}s' : '获取'),
                    ),
                  ],
                ),
              const SizedBox(height: 32),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: auth.loading ? null : _login,
                  child: auth.loading
                      ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                      : const Text('登 录', style: TextStyle(fontSize: 16)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
