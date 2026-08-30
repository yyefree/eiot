import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';

class HomeProvider extends ChangeNotifier {
  List<Home> _homes = [];
  Home? _currentHome;
  List<Room> _rooms = [];
  bool _loading = false;

  List<Home> get homes => _homes;
  Home? get currentHome => _currentHome;
  List<Room> get rooms => _rooms;
  bool get loading => _loading;

  Future<void> loadHomes() async {
    _loading = true;
    notifyListeners();
    final resp = await ApiClient.get('/home');
    if (resp.ok && resp.data != null) {
      final data = resp.data;
      final list = (data is Map<String, dynamic> ? data['list'] : data) as List?;
      _homes = list?.whereType<Map<String, dynamic>>().map((e) => Home.fromJson(e)).toList() ?? [];
      if (_homes.isNotEmpty && _currentHome == null) {
        _currentHome = _homes.firstWhere((h) => h.isDefault, orElse: () => _homes.first);
      }
    }
    _loading = false;
    notifyListeners();
  }

  void switchHome(Home home) {
    _currentHome = home;
    notifyListeners();
    loadRooms();
  }

  Future<void> loadRooms() async {
    if (_currentHome == null) return;
    final resp = await ApiClient.getRoomList(_currentHome!.id);
    if (resp.ok && resp.data != null) {
      final data = resp.data;
      final list = (data is Map<String, dynamic> ? data['list'] : data) as List?;
      _rooms = list?.whereType<Map<String, dynamic>>().map((e) => Room.fromJson(e)).toList() ?? [];
    }
    notifyListeners();
  }

  Future<bool> createHome(String name, String address) async {
    final resp = await ApiClient.post('/home', {'name': name, 'address': address});
    if (resp.ok) {
      await loadHomes();
      return true;
    }
    return false;
  }

  Future<bool> updateHome(int id, String name, String address) async {
    final resp = await ApiClient.put('/home/$id', {'name': name, 'address': address});
    if (resp.ok) {
      await loadHomes();
      return true;
    }
    return false;
  }

  Future<bool> deleteHome(int id) async {
    final resp = await ApiClient.delete('/home/$id');
    if (resp.ok) {
      if (_currentHome?.id == id) _currentHome = null;
      await loadHomes();
      return true;
    }
    return false;
  }

  Future<bool> createRoom(String name, String icon) async {
    if (_currentHome == null) return false;
    final resp = await ApiClient.createRoom(_currentHome!.id, {'name': name, 'icon': icon});
    if (resp.ok) {
      await loadRooms();
      return true;
    }
    return false;
  }

  Future<bool> updateRoom(int roomId, String name, String icon) async {
    if (_currentHome == null) return false;
    final resp = await ApiClient.updateRoom(_currentHome!.id, roomId, {'name': name, 'icon': icon});
    if (resp.ok) {
      await loadRooms();
      return true;
    }
    return false;
  }

  Future<bool> deleteRoom(int roomId) async {
    if (_currentHome == null) return false;
    final resp = await ApiClient.deleteRoom(_currentHome!.id, roomId);
    if (resp.ok) {
      await loadRooms();
      return true;
    }
    return false;
  }

  Future<List<DeviceItem>> getRoomDevices(int roomId) async {
    if (_currentHome == null) return [];
    final resp = await ApiClient.getRoomDevices(_currentHome!.id, roomId);
    if (resp.ok && resp.data != null) {
      final data = resp.data;
      final list = (data is Map<String, dynamic> ? data['list'] : data) as List?;
      return list?.whereType<Map<String, dynamic>>().map((e) => DeviceItem.fromJson(e)).toList() ?? [];
    }
    return [];
  }

  Future<bool> addDeviceToRoom(int roomId, int deviceId) async {
    if (_currentHome == null) return false;
    final resp = await ApiClient.addDeviceToRoom(_currentHome!.id, roomId, {'device_id': deviceId});
    return resp.ok;
  }

  Future<bool> removeDeviceFromRoom(int roomId, int deviceId) async {
    if (_currentHome == null) return false;
    final resp = await ApiClient.removeDeviceFromRoom(_currentHome!.id, roomId, deviceId);
    return resp.ok;
  }
}
