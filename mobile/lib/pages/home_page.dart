import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../providers/home_provider.dart';
import '../api/api_client.dart';
import '../models/models.dart';
import 'device_detail_page.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  List<DeviceItem> _devices = [];
  bool _loading = true;
  int? _selectedRoomId;

  @override
  void initState() {
    super.initState();
    _loadDevices();
  }

  Future<void> _loadDevices() async {
    setState(() => _loading = true);
    final resp = await ApiClient.get('/device', params: {'page': 1, 'size': 100});
    if (mounted) {
      if (resp.ok && resp.data != null) {
        final data = resp.data as Map<String, dynamic>;
        final list = (data['list'] as List?)
                ?.whereType<Map<String, dynamic>>()
                .map((e) => DeviceItem.fromJson(e))
                .toList() ??
            [];
        setState(() {
          _devices = list;
          _loading = false;
        });
      } else {
        setState(() => _loading = false);
      }
    }
  }

  List<DeviceItem> get _filteredDevices {
    if (_selectedRoomId == null) return _devices;
    return _devices.where((d) => d.roomId == _selectedRoomId).toList();
  }

  Future<void> _toggleDevice(DeviceItem device) async {
    await ApiClient.post('/device/${device.id}/control', {'power_switch': true});
    _loadDevices();
  }

  IconData _deviceIcon(String productName) {
    final name = productName.toLowerCase();
    if (name.contains('灯') || name.contains('light') || name.contains('lamp')) return Icons.lightbulb_outline;
    if (name.contains('空调') || name.contains('air')) return Icons.ac_unit;
    if (name.contains('插座') || name.contains('socket') || name.contains('plug')) return Icons.power;
    if (name.contains('传感器') || name.contains('sensor')) return Icons.sensors;
    if (name.contains('门锁') || name.contains('lock')) return Icons.lock_outline;
    if (name.contains('窗帘') || name.contains('curtain')) return Icons.curtains;
    if (name.contains('摄像头') || name.contains('camera')) return Icons.videocam_outlined;
    return Icons.devices_other;
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final homeProvider = context.watch<HomeProvider>();
    final user = auth.user;

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      body: RefreshIndicator(
        onRefresh: () async {
          await _loadDevices();
          await homeProvider.loadHomes();
        },
        child: CustomScrollView(
          slivers: [
            SliverAppBar(
              floating: true,
              backgroundColor: Colors.white,
              title: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    homeProvider.currentHome?.name ?? '我的家',
                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(width: 4),
                  const Icon(Icons.swap_horiz, size: 18, color: Color(0xFF007DFF)),
                ],
              ),
              actions: [
                IconButton(
                  icon: const Icon(Icons.qr_code_scanner, color: Color(0xFF333333)),
                  onPressed: () {},
                ),
              ],
            ),
            SliverToBoxAdapter(
              child: Container(
                color: Colors.white,
                padding: const EdgeInsets.only(bottom: 12),
                child: SizedBox(
                  height: 40,
                  child: ListView(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    children: [
                      _roomChip('全部', _selectedRoomId == null, () {
                        setState(() => _selectedRoomId = null);
                      }),
                      ...homeProvider.rooms.map((room) => _roomChip(
                        room.name,
                        _selectedRoomId == room.id,
                        () => setState(() => _selectedRoomId = room.id),
                      )),
                    ],
                  ),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: _buildSceneShortcuts(),
            ),
            if (_loading)
              const SliverFillRemaining(
                child: Center(child: CircularProgressIndicator()),
              )
            else if (_filteredDevices.isEmpty)
              SliverFillRemaining(
                child: _buildEmptyState(),
              )
            else
              SliverPadding(
                padding: const EdgeInsets.all(12),
                sliver: SliverGrid(
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    mainAxisSpacing: 12,
                    crossAxisSpacing: 12,
                    childAspectRatio: 1.0,
                  ),
                  delegate: SliverChildBuilderDelegate(
                    (context, index) => _buildDeviceCard(_filteredDevices[index]),
                    childCount: _filteredDevices.length,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _roomChip(String label, bool selected, VoidCallback onTap) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFF007DFF) : const Color(0xFFF0F0F0),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: selected ? Colors.white : const Color(0xFF666666),
              fontSize: 13,
              fontWeight: selected ? FontWeight.w500 : FontWeight.normal,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSceneShortcuts() {
    return Container(
      color: Colors.white,
      margin: const EdgeInsets.only(top: 8),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('场景快捷', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          SizedBox(
            height: 80,
            child: ListView(
              scrollDirection: Axis.horizontal,
              children: [
                _sceneShortcutCard('回家模式', Icons.home, const Color(0xFF4CAF50)),
                const SizedBox(width: 12),
                _sceneShortcutCard('离家模式', Icons.exit_to_app, const Color(0xFFFF9800)),
                const SizedBox(width: 12),
                _sceneShortcutCard('睡眠模式', Icons.bedtime, const Color(0xFF9C27B0)),
                const SizedBox(width: 12),
                _sceneShortcutCard('起床模式', Icons.alarm, const Color(0xFF2196F3)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _sceneShortcutCard(String name, IconData icon, Color color) {
    return GestureDetector(
      onTap: () {},
      child: Container(
        width: 120,
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: color.withOpacity(0.1),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, color: color, size: 28),
            const SizedBox(height: 8),
            Text(name, style: TextStyle(color: color, fontSize: 13, fontWeight: FontWeight.w500)),
          ],
        ),
      ),
    );
  }

  Widget _buildDeviceCard(DeviceItem device) {
    return GestureDetector(
      onTap: () async {
        await Navigator.push(context, MaterialPageRoute(builder: (_) => DeviceDetailPage(deviceId: device.id)));
        _loadDevices();
      },
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Icon(_deviceIcon(device.productName), size: 32, color: const Color(0xFF007DFF)),
                Row(
                  children: [
                    Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: device.online ? const Color(0xFF4CAF50) : const Color(0xFFCCCCCC),
                      ),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      device.online ? '在线' : '离线',
                      style: TextStyle(
                        fontSize: 11,
                        color: device.online ? const Color(0xFF4CAF50) : const Color(0xFFCCCCCC),
                      ),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 8),
            Align(
              alignment: Alignment.centerLeft,
              child: Text(
                device.name,
                style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Align(
              alignment: Alignment.centerLeft,
              child: Text(
                device.productName,
                style: const TextStyle(fontSize: 11, color: Color(0xFF999999)),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Align(
              alignment: Alignment.centerRight,
              child: Switch(
                value: device.online,
                onChanged: device.online ? (_) => _toggleDevice(device) : null,
                activeColor: const Color(0xFF007DFF),
                materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.devices_other, size: 80, color: const Color(0xFFDDDDDD)),
          const SizedBox(height: 16),
          const Text('暂无设备', style: TextStyle(fontSize: 16, color: Color(0xFF999999))),
          const SizedBox(height: 8),
          const Text('点击下方 + 添加设备', style: TextStyle(fontSize: 13, color: Color(0xFFCCCCCC))),
        ],
      ),
    );
  }
}
