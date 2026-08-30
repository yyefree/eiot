import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/home_provider.dart';
import '../models/models.dart';

class RoomManagerPage extends StatefulWidget {
  final Home home;
  const RoomManagerPage({super.key, required this.home});

  @override
  State<RoomManagerPage> createState() => _RoomManagerPageState();
}

class _RoomManagerPageState extends State<RoomManagerPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<HomeProvider>().loadRooms();
    });
  }

  @override
  Widget build(BuildContext context) {
    final homeProvider = context.watch<HomeProvider>();

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: Text('${widget.home.name} · 房间管理'),
      ),
      body: homeProvider.rooms.isEmpty
          ? _buildEmptyState()
          : RefreshIndicator(
              onRefresh: homeProvider.loadRooms,
              child: GridView.builder(
                padding: const EdgeInsets.all(12),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  mainAxisSpacing: 12,
                  crossAxisSpacing: 12,
                  childAspectRatio: 1.2,
                ),
                itemCount: homeProvider.rooms.length,
                itemBuilder: (context, index) => _buildRoomCard(homeProvider.rooms[index], homeProvider),
              ),
            ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCreateRoomDialog(homeProvider),
        backgroundColor: const Color(0xFF007DFF),
        child: const Icon(Icons.add, color: Colors.white),
      ),
    );
  }

  Widget _buildRoomCard(Room room, HomeProvider provider) {
    return Card(
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () {
          _showRoomDetail(room, provider);
        },
        onLongPress: () {
          _showRoomOptions(room, provider);
        },
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                _roomIcon(room.icon),
                size: 36,
                color: const Color(0xFF007DFF),
              ),
              const SizedBox(height: 12),
              Text(
                room.name,
                style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 4),
              Text(
                '${room.deviceCount} 个设备',
                style: const TextStyle(fontSize: 12, color: Color(0xFF999999)),
              ),
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
          Icon(Icons.meeting_room, size: 80, color: const Color(0xFFDDDDDD)),
          const SizedBox(height: 16),
          const Text('暂无房间', style: TextStyle(fontSize: 16, color: Color(0xFF999999))),
          const SizedBox(height: 8),
          const Text('点击右下角 + 添加房间', style: TextStyle(fontSize: 13, color: Color(0xFFCCCCCC))),
        ],
      ),
    );
  }

  void _showCreateRoomDialog(HomeProvider provider) {
    final nameCtrl = TextEditingController();
    String selectedIcon = 'meeting_room';
    final icons = [
      'meeting_room', 'living', 'bed_child', 'restaurant', 'bathtub',
      'garage', 'yard', 'light', 'meeting_room', 'store',
    ];

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: const Text('添加房间'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: nameCtrl,
                decoration: const InputDecoration(labelText: '房间名称', hintText: '客厅'),
              ),
              const SizedBox(height: 16),
              const Align(
                alignment: Alignment.centerLeft,
                child: Text('选择图标', style: TextStyle(fontSize: 13, color: Color(0xFF999999))),
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: icons.map((icon) => GestureDetector(
                  onTap: () => setDialogState(() => selectedIcon = icon),
                  child: Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: selectedIcon == icon
                          ? const Color(0xFF007DFF).withOpacity(0.1)
                          : const Color(0xFFF5F5F5),
                      borderRadius: BorderRadius.circular(10),
                      border: selectedIcon == icon
                          ? Border.all(color: const Color(0xFF007DFF), width: 2)
                          : null,
                    ),
                    child: Icon(_roomIcon(icon), size: 22),
                  ),
                )).toList(),
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
            TextButton(
              onPressed: () async {
                if (nameCtrl.text.isNotEmpty) {
                  await provider.createRoom(nameCtrl.text, selectedIcon);
                  if (ctx.mounted) Navigator.pop(ctx);
                }
              },
              child: const Text('创建'),
            ),
          ],
        ),
      ),
    );
  }

  void _showRoomDetail(Room room, HomeProvider provider) {
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
                  width: 40, height: 4,
                  decoration: BoxDecoration(color: const Color(0xFFDDDDDD), borderRadius: BorderRadius.circular(2)),
                ),
              ),
              const SizedBox(height: 20),
              Row(
                children: [
                  Icon(_roomIcon(room.icon), size: 28, color: const Color(0xFF007DFF)),
                  const SizedBox(width: 12),
                  Text(room.name, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600)),
                ],
              ),
              const SizedBox(height: 20),
              const Divider(),
              FutureBuilder<List<DeviceItem>>(
                future: provider.getRoomDevices(room.id),
                builder: (context, snapshot) {
                  if (snapshot.connectionState == ConnectionState.waiting) {
                    return const Center(child: CircularProgressIndicator());
                  }
                  final devices = snapshot.data ?? [];
                  if (devices.isEmpty) {
                    return const Padding(
                      padding: EdgeInsets.all(40),
                      child: Center(
                        child: Text('该房间暂无设备', style: TextStyle(color: Color(0xFF999999))),
                      ),
                    );
                  }
                  return Column(
                    children: devices.map((d) => ListTile(
                      leading: Icon(
                        d.online ? Icons.wifi : Icons.wifi_off,
                        color: d.online ? const Color(0xFF4CAF50) : const Color(0xFFCCCCCC),
                      ),
                      title: Text(d.name),
                      subtitle: Text(d.productName, style: const TextStyle(fontSize: 12)),
                    )).toList(),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showRoomOptions(Room room, HomeProvider provider) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.edit),
              title: const Text('编辑房间'),
              onTap: () {
                Navigator.pop(context);
                _showEditRoomDialog(room, provider);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('删除房间', style: TextStyle(color: Colors.red)),
              onTap: () {
                Navigator.pop(context);
                _confirmDeleteRoom(room, provider);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showEditRoomDialog(Room room, HomeProvider provider) {
    final nameCtrl = TextEditingController(text: room.name);
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('编辑房间'),
        content: TextField(controller: nameCtrl, decoration: const InputDecoration(labelText: '房间名称')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              if (nameCtrl.text.isNotEmpty) {
                await provider.updateRoom(room.id, nameCtrl.text, room.icon);
                if (ctx.mounted) Navigator.pop(ctx);
              }
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _confirmDeleteRoom(Room room, HomeProvider provider) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除房间'),
        content: Text('确定要删除"${room.name}"吗？'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              await provider.deleteRoom(room.id);
              if (ctx.mounted) Navigator.pop(ctx);
            },
            child: const Text('删除', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }

  IconData _roomIcon(String iconName) {
    switch (iconName) {
      case 'living': return Icons.living;
      case 'bed_child': return Icons.bed;
      case 'restaurant': return Icons.restaurant;
      case 'bathtub': return Icons.bathtub;
      case 'garage': return Icons.garage;
      case 'yard': return Icons.yard;
      case 'light': return Icons.lightbulb;
      case 'meeting_room': return Icons.meeting_room;
      case 'store': return Icons.store;
      default: return Icons.meeting_room;
    }
  }
}
