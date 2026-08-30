import 'package:flutter/material.dart';
import '../api/api_client.dart';
import '../models/models.dart';

class MessageProvider extends ChangeNotifier {
  List<Message> _messages = [];
  bool _loading = false;
  String _filter = 'all';
  int _unreadCount = 0;

  List<Message> get messages => _messages;
  bool get loading => _loading;
  String get filter => _filter;
  int get unreadCount => _unreadCount;

  List<Message> get filteredMessages {
    if (_filter == 'all') return _messages;
    return _messages.where((m) => m.type == _filter).toList();
  }

  void setFilter(String filter) {
    _filter = filter;
    notifyListeners();
  }

  Future<void> loadMessages() async {
    _loading = true;
    notifyListeners();
    final resp = await ApiClient.get('/message', params: {'page': 1, 'size': 100});
    if (resp.ok && resp.data != null) {
      final data = resp.data;
      final list = (data is Map<String, dynamic> ? data['list'] : data) as List?;
      _messages = list?.whereType<Map<String, dynamic>>().map((e) => Message.fromJson(e)).toList() ?? [];
      _unreadCount = _messages.where((m) => !m.read).length;
    }
    _loading = false;
    notifyListeners();
  }

  Future<void> markAsRead(int id) async {
    await ApiClient.post('/message/$id/read', {});
    final idx = _messages.indexWhere((m) => m.id == id);
    if (idx >= 0) {
      _messages[idx] = Message(
        id: _messages[idx].id,
        type: _messages[idx].type,
        title: _messages[idx].title,
        content: _messages[idx].content,
        read: true,
        createdAt: _messages[idx].createdAt,
      );
      _unreadCount = _messages.where((m) => !m.read).length;
      notifyListeners();
    }
  }

  Future<void> markAllAsRead() async {
    await ApiClient.post('/message/read-all', {});
    _messages = _messages.map((m) => Message(
      id: m.id,
      type: m.type,
      title: m.title,
      content: m.content,
      read: true,
      createdAt: m.createdAt,
    )).toList();
    _unreadCount = 0;
    notifyListeners();
  }

  Future<void> deleteMessage(int id) async {
    await ApiClient.delete('/message/$id');
    _messages.removeWhere((m) => m.id == id);
    _unreadCount = _messages.where((m) => !m.read).length;
    notifyListeners();
  }
}
