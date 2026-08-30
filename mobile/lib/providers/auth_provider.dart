import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';

class AuthProvider extends ChangeNotifier {
  UserInfo? _user;
  bool _loading = false;
  bool _initialized = false;

  UserInfo? get user => _user;
  bool get isLoggedIn => ApiClient.token != null;
  bool get loading => _loading;
  bool get initialized => _initialized;

  Future<void> init() async {
    try {
      await ApiClient.init();
      if (ApiClient.token != null) {
        await loadUser();
      }
    } catch (_) {}
    _initialized = true;
    notifyListeners();
  }

  Future<void> loadUser() async {
    try {
      final resp = await ApiClient.get('/user/info');
      if (resp.ok && resp.data != null && resp.data is Map<String, dynamic>) {
        _user = UserInfo.fromJson(resp.data as Map<String, dynamic>);
        notifyListeners();
      }
    } catch (_) {}
  }

  Future<bool> login(String phone, String password) async {
    _loading = true;
    notifyListeners();
    try {
      final resp = await ApiClient.post('/auth/login', {'phone': phone, 'password': password});
      _loading = false;
      if (resp.ok && resp.data != null && resp.data is Map<String, dynamic>) {
        final data = resp.data as Map<String, dynamic>;
        final token = data['token'] as String?;
        if (token != null) {
          await ApiClient.saveToken(token);
        }
        if (data['user'] is Map<String, dynamic>) {
          _user = UserInfo.fromJson(data['user'] as Map<String, dynamic>);
        }
        notifyListeners();
        return true;
      }
      notifyListeners();
      return false;
    } catch (_) {
      _loading = false;
      notifyListeners();
      return false;
    }
  }

  Future<bool> sendCode(String phone) async {
    try {
      final resp = await ApiClient.post('/auth/send-code', {'phone': phone});
      return resp.ok;
    } catch (_) {
      return false;
    }
  }

  Future<bool> loginByCode(String phone, String code) async {
    _loading = true;
    notifyListeners();
    try {
      final resp = await ApiClient.post('/auth/login-code', {'phone': phone, 'code': code});
      _loading = false;
      if (resp.ok && resp.data != null && resp.data is Map<String, dynamic>) {
        final data = resp.data as Map<String, dynamic>;
        final token = data['token'] as String?;
        if (token != null) {
          await ApiClient.saveToken(token);
        }
        if (data['user'] is Map<String, dynamic>) {
          _user = UserInfo.fromJson(data['user'] as Map<String, dynamic>);
        }
        notifyListeners();
        return true;
      }
      notifyListeners();
      return false;
    } catch (_) {
      _loading = false;
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    await ApiClient.clearToken();
    _user = null;
    notifyListeners();
  }
}
