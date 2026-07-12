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
    await ApiClient.init();
    if (ApiClient.token != null) {
      await loadUser();
    }
    _initialized = true;
    notifyListeners();
  }

  Future<void> loadUser() async {
    final resp = await ApiClient.get('/user/info');
    if (resp.ok && resp.data != null) {
      _user = UserInfo.fromJson(resp.data as Map<String, dynamic>);
      notifyListeners();
    }
  }

  Future<bool> login(String phone, String password) async {
    _loading = true;
    notifyListeners();
    final resp = await ApiClient.post('/auth/login', {'phone': phone, 'password': password});
    _loading = false;
    if (resp.ok && resp.data != null) {
      final data = resp.data as Map<String, dynamic>;
      await ApiClient.saveToken(data['token'] as String);
      _user = UserInfo.fromJson(data['user'] as Map<String, dynamic>);
      notifyListeners();
      return true;
    }
    notifyListeners();
    return false;
  }

  Future<String?> sendCode(String phone) async {
    final resp = await ApiClient.post('/auth/send-code', {'phone': phone});
    if (resp.ok && resp.data != null) {
      return resp.data['code'] as String?;
    }
    return null;
  }

  Future<bool> loginByCode(String phone, String code) async {
    _loading = true;
    notifyListeners();
    final resp = await ApiClient.post('/auth/login-code', {'phone': phone, 'code': code});
    _loading = false;
    if (resp.ok && resp.data != null) {
      final data = resp.data as Map<String, dynamic>;
      await ApiClient.saveToken(data['token'] as String);
      _user = UserInfo.fromJson(data['user'] as Map<String, dynamic>);
      notifyListeners();
      return true;
    }
    notifyListeners();
    return false;
  }

  Future<void> logout() async {
    await ApiClient.clearToken();
    _user = null;
    notifyListeners();
  }
}
