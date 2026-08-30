import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';

class SceneProvider extends ChangeNotifier {
  List<Scene> _manualScenes = [];
  List<Scene> _autoScenes = [];
  bool _loading = false;

  List<Scene> get manualScenes => _manualScenes;
  List<Scene> get autoScenes => _autoScenes;
  bool get loading => _loading;

  Future<void> loadScenes() async {
    _loading = true;
    notifyListeners();
    final resp = await ApiClient.getSceneList();
    if (resp.ok && resp.data != null) {
      final data = resp.data;
      final list = (data is Map<String, dynamic> ? data['list'] : data) as List?;
      final all = list?.whereType<Map<String, dynamic>>().map((e) => Scene.fromJson(e)).toList() ?? [];
      _manualScenes = all.where((s) => s.type == 'manual').toList();
      _autoScenes = all.where((s) => s.type == 'auto').toList();
    }
    _loading = false;
    notifyListeners();
  }

  Future<bool> createScene(Map<String, dynamic> body) async {
    final resp = await ApiClient.post('/scene', body);
    if (resp.ok) {
      await loadScenes();
      return true;
    }
    return false;
  }

  Future<bool> updateScene(int id, Map<String, dynamic> body) async {
    final resp = await ApiClient.put('/scene/$id', body);
    if (resp.ok) {
      await loadScenes();
      return true;
    }
    return false;
  }

  Future<bool> deleteScene(int id) async {
    final resp = await ApiClient.delete('/scene/$id');
    if (resp.ok) {
      await loadScenes();
      return true;
    }
    return false;
  }

  Future<bool> runScene(int id) async {
    final resp = await ApiClient.post('/scene/$id/run', {});
    return resp.ok;
  }

  Future<bool> toggleScene(int id, bool enabled) async {
    final resp = await ApiClient.toggleScene(id, enabled);
    if (resp.ok) {
      await loadScenes();
      return true;
    }
    return false;
  }
}
