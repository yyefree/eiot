import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

class ApiResponse {
  final int code;
  final String msg;
  final dynamic data;
  ApiResponse({required this.code, required this.msg, this.data});
  bool get ok => code == 0;
}

class ApiClient {
  static String _baseUrl = 'http://localhost:8080/api';
  static String? _token;

  static String get baseUrl => _baseUrl;

  static Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('eiot_token');
    _baseUrl = prefs.getString('eiot_base_url') ?? 'http://localhost:8080/api';
  }

  static Future<void> setBaseUrl(String url) async {
    _baseUrl = url;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('eiot_base_url', url);
  }

  static Future<void> saveToken(String token) async {
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('eiot_token', token);
  }

  static Future<void> clearToken() async {
    _token = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('eiot_token');
  }

  static String? get token => _token;

  static Map<String, String> get _headers => {
    'Content-Type': 'application/json',
    if (_token != null) 'Authorization': 'Bearer $_token',
  };

  static Future<ApiResponse> get(String path, {Map<String, dynamic>? params}) async {
    var url = '$baseUrl$path';
    if (params != null && params.isNotEmpty) {
      final q = params.entries.where((e) => e.value != null).map((e) => '${Uri.encodeComponent(e.key)}=${Uri.encodeComponent('${e.value}')}'  ).join('&');
      url += '?$q';
    }
    try {
      final resp = await http.get(Uri.parse(url), headers: _headers).timeout(const Duration(seconds: 15));
      return _parse(resp);
    } catch (e) {
      return ApiResponse(code: -1, msg: '网络连接失败', data: null);
    }
  }

  static Future<ApiResponse> post(String path, Map<String, dynamic> body) async {
    try {
      final resp = await http.post(Uri.parse('$baseUrl$path'), headers: _headers, body: jsonEncode(body)).timeout(const Duration(seconds: 15));
      return _parse(resp);
    } catch (e) {
      return ApiResponse(code: -1, msg: '网络连接失败', data: null);
    }
  }

  static Future<ApiResponse> put(String path, Map<String, dynamic> body) async {
    try {
      final resp = await http.put(Uri.parse('$baseUrl$path'), headers: _headers, body: jsonEncode(body)).timeout(const Duration(seconds: 15));
      return _parse(resp);
    } catch (e) {
      return ApiResponse(code: -1, msg: '网络连接失败', data: null);
    }
  }

  static Future<ApiResponse> delete(String path) async {
    try {
      final resp = await http.delete(Uri.parse('$baseUrl$path'), headers: _headers).timeout(const Duration(seconds: 15));
      return _parse(resp);
    } catch (e) {
      return ApiResponse(code: -1, msg: '网络连接失败', data: null);
    }
  }

  static ApiResponse _parse(http.Response resp) {
    try {
      final json = jsonDecode(resp.body) as Map<String, dynamic>;
      return ApiResponse(
        code: json['code'] as int? ?? -1,
        msg: json['msg'] as String? ?? '',
        data: json['data'],
      );
    } catch (_) {
      return ApiResponse(code: -1, msg: '网络错误 (${resp.statusCode})', data: null);
    }
  }
}
