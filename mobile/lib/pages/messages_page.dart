import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/message_provider.dart';
import '../models/models.dart';

class MessagesPage extends StatefulWidget {
  const MessagesPage({super.key});

  @override
  State<MessagesPage> createState() => _MessagesPageState();
}

class _MessagesPageState extends State<MessagesPage> {
  final _filters = [
    {'key': 'all', 'label': '全部'},
    {'key': 'system', 'label': '系统'},
    {'key': 'device', 'label': '设备'},
    {'key': 'scene', 'label': '场景'},
  ];

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<MessageProvider>().loadMessages();
    });
  }

  @override
  Widget build(BuildContext context) {
    final messageProvider = context.watch<MessageProvider>();

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: const Text('消息中心'),
        actions: [
          if (messageProvider.unreadCount > 0)
            TextButton(
              onPressed: () => messageProvider.markAllAsRead(),
              child: const Text('全部已读', style: TextStyle(color: Color(0xFF007DFF))),
            ),
        ],
      ),
      body: Column(
        children: [
          Container(
            color: Colors.white,
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: SizedBox(
              height: 36,
              child: ListView.separated(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: _filters.length,
                separatorBuilder: (_, __) => const SizedBox(width: 8),
                itemBuilder: (context, index) {
                  final f = _filters[index];
                  final selected = messageProvider.filter == f['key'];
                  return GestureDetector(
                    onTap: () => messageProvider.setFilter(f['key']!),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                      decoration: BoxDecoration(
                        color: selected ? const Color(0xFF007DFF) : const Color(0xFFF0F0F0),
                        borderRadius: BorderRadius.circular(18),
                      ),
                      child: Text(
                        f['label']!,
                        style: TextStyle(
                          color: selected ? Colors.white : const Color(0xFF666666),
                          fontSize: 13,
                          fontWeight: selected ? FontWeight.w500 : FontWeight.normal,
                        ),
                      ),
                    ),
                  );
                },
              ),
            ),
          ),
          Expanded(
            child: messageProvider.loading
                ? const Center(child: CircularProgressIndicator())
                : messageProvider.filteredMessages.isEmpty
                    ? _buildEmptyState()
                    : RefreshIndicator(
                        onRefresh: messageProvider.loadMessages,
                        child: ListView.builder(
                          padding: const EdgeInsets.all(12),
                          itemCount: messageProvider.filteredMessages.length,
                          itemBuilder: (context, index) {
                            final msg = messageProvider.filteredMessages[index];
                            return Dismissible(
                              key: Key('msg_${msg.id}'),
                              direction: DismissDirection.endToStart,
                              background: Container(
                                alignment: Alignment.centerRight,
                                padding: const EdgeInsets.only(right: 20),
                                color: Colors.red,
                                child: const Icon(Icons.delete, color: Colors.white),
                              ),
                              onDismissed: (_) => messageProvider.deleteMessage(msg.id),
                              child: _buildMessageCard(msg, messageProvider),
                            );
                          },
                        ),
                      ),
          ),
        ],
      ),
    );
  }

  Widget _buildMessageCard(Message msg, MessageProvider provider) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () {
          if (!msg.read) provider.markAsRead(msg.id);
          _showMessageDetail(msg);
        },
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: _messageColor(msg.type).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(_messageIcon(msg.type), color: _messageColor(msg.type), size: 22),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            msg.title,
                            style: TextStyle(
                              fontSize: 15,
                              fontWeight: msg.read ? FontWeight.normal : FontWeight.w600,
                            ),
                          ),
                        ),
                        if (!msg.read)
                          Container(
                            width: 8,
                            height: 8,
                            decoration: const BoxDecoration(
                              shape: BoxShape.circle,
                              color: Color(0xFFFF4444),
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Text(
                      msg.content,
                      style: const TextStyle(fontSize: 13, color: Color(0xFF999999)),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 6),
                    Text(
                      msg.createdAt,
                      style: const TextStyle(fontSize: 11, color: Color(0xFFCCCCCC)),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showMessageDetail(Message msg) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.5,
        minChildSize: 0.3,
        maxChildSize: 0.8,
        expand: false,
        builder: (context, scrollController) => Container(
          padding: const EdgeInsets.all(20),
          child: ListView(
            controller: scrollController,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: const Color(0xFFDDDDDD),
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 20),
              Row(
                children: [
                  Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: _messageColor(msg.type).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(_messageIcon(msg.type), color: _messageColor(msg.type), size: 26),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(msg.title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
                        const SizedBox(height: 4),
                        Text(msg.createdAt, style: const TextStyle(fontSize: 12, color: Color(0xFF999999))),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 20),
              const Divider(),
              const SizedBox(height: 16),
              Text(msg.content, style: const TextStyle(fontSize: 15, height: 1.6)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.notifications_none, size: 80, color: const Color(0xFFDDDDDD)),
          const SizedBox(height: 16),
          const Text('暂无消息', style: TextStyle(fontSize: 16, color: Color(0xFF999999))),
        ],
      ),
    );
  }

  IconData _messageIcon(String type) {
    switch (type) {
      case 'device': return Icons.devices_other;
      case 'scene': return Icons.auto_awesome;
      case 'security': return Icons.security;
      default: return Icons.info_outline;
    }
  }

  Color _messageColor(String type) {
    switch (type) {
      case 'device': return const Color(0xFF007DFF);
      case 'scene': return const Color(0xFF4CAF50);
      case 'security': return const Color(0xFFFF9800);
      default: return const Color(0xFF666666);
    }
  }
}
