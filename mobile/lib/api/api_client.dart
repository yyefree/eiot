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
  static String _baseUrl = 'http://192.168.1.9:8080/api';
  static String? _token;

  static String get baseUrl => _baseUrl;

  static Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('eiot_token');
    _baseUrl = prefs.getString('eiot_base_url') ?? _baseUrl;
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

  static Future<ApiResponse> getHomeList() => get('/home');
  static Future<ApiResponse> createHome(Map<String, dynamic> body) => post('/home', body);
  static Future<ApiResponse> updateHome(int id, Map<String, dynamic> body) => put('/home/$id', body);
  static Future<ApiResponse> deleteHome(int id) => delete('/home/$id');

  static Future<ApiResponse> getRoomList(int homeId) => get('/room', params: {'home_id': homeId});
  static Future<ApiResponse> createRoom(int homeId, Map<String, dynamic> body) => post('/room', {...body, 'home_id': homeId});
  static Future<ApiResponse> updateRoom(int homeId, int roomId, Map<String, dynamic> body) => put('/room/$roomId', body);
  static Future<ApiResponse> deleteRoom(int homeId, int roomId) => delete('/room/$roomId');

  static Future<ApiResponse> getRoomDevices(int homeId, int roomId) => get('/room/$roomId/device');
  static Future<ApiResponse> addDeviceToRoom(int homeId, int roomId, Map<String, dynamic> body) => post('/room/$roomId/device', body);
  static Future<ApiResponse> removeDeviceFromRoom(int homeId, int roomId, int deviceId) => delete('/room/$roomId/device/$deviceId');

  static Future<ApiResponse> getSceneList({int? homeId}) => get('/scene', params: homeId != null ? {'home_id': homeId} : null);
  static Future<ApiResponse> createScene(Map<String, dynamic> body) => post('/scene', body);
  static Future<ApiResponse> updateScene(int id, Map<String, dynamic> body) => put('/scene/$id', body);
  static Future<ApiResponse> deleteScene(int id) => delete('/scene/$id');
  static Future<ApiResponse> runScene(int id) => post('/scene/$id/run', {});
  static Future<ApiResponse> toggleScene(int id, bool enabled) => put('/scene/$id/toggle', {'enabled': enabled});

  static Future<ApiResponse> getMessageList({int page = 1, int size = 50}) => get('/message', params: {'page': page, 'size': size});
  static Future<ApiResponse> markMessageRead(int id) => put('/message/$id/read', {});
  static Future<ApiResponse> markAllMessagesRead() => put('/message/read-all', {});
  static Future<ApiResponse> deleteMessage(int id) => delete('/message/$id');

  static Future<ApiResponse> getOTAFirmwareList() => get('/admin/ota');
  static Future<ApiResponse> getOTAFirmwareDetail(int id) => get('/admin/ota/$id');

  static Future<ApiResponse> getUserInfo() => get('/user/info');
  static Future<ApiResponse> updateUserInfo(Map<String, dynamic> body) => put('/user/info', body);

  static Future<ApiResponse> getDeviceDataHistory(String deviceSn, {String? property, int limit = 100}) =>
      get('/device/data/$deviceSn', params: { if (property != null) 'property': property, 'limit': limit });

  static Future<ApiResponse> getDeviceEvents(int deviceId, {int page = 1, int size = 20, String? eventId}) async {
    var path = '/device/$deviceId/event?page=$page&size=$size';
    if (eventId != null && eventId.isNotEmpty) path += '&event_id=$eventId';
    return get(path);
  }

  static Future<ApiResponse> reportDeviceEvent(int deviceId, String eventId, String eventName, Map<String, dynamic> output) async {
    return post('/device/$deviceId/event', {
      'event_id': eventId,
      'event_name': eventName,
      'output': output,
    });
  }

  static Future<ApiResponse> getDeviceServices(int deviceId, {int page = 1, int size = 20, String? serviceId}) async {
    var path = '/device/$deviceId/service?page=$page&size=$size';
    if (serviceId != null && serviceId.isNotEmpty) path += '&service_id=$serviceId';
    return get(path);
  }

  static Future<ApiResponse> invokeDeviceService(int deviceId, String serviceId, String serviceName, Map<String, dynamic> input) async {
    return post('/device/$deviceId/service', {
      'service_id': serviceId,
      'service_name': serviceName,
      'input': input,
    });
  }

  static Future<ApiResponse> getDeviceShadow(int deviceId) async {
    return get('/device/$deviceId/shadow');
  }

  static Future<ApiResponse> updateDeviceShadow(int deviceId, Map<String, dynamic> desired) async {
    return put('/device/$deviceId/shadow', {'desired': desired});
  }

  static Future<ApiResponse> healthCheck() => get('/health');
}
