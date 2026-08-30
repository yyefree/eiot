import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/scene_provider.dart';
import '../models/models.dart';

class SceneCreatePage extends StatefulWidget {
  const SceneCreatePage({super.key});

  @override
  State<SceneCreatePage> createState() => _SceneCreatePageState();
}

class _SceneCreatePageState extends State<SceneCreatePage> {
  final _nameCtrl = TextEditingController();
  String _type = 'manual';
  String _selectedIcon = 'play_circle';
  final List<SceneCondition> _conditions = [];
  final List<SceneAction> _actions = [];

  final _icons = [
    'play_circle', 'home', 'exit_to_app', 'bedtime', 'alarm',
    'lightbulb', 'security', 'temperature', 'curtains',
  ];

  @override
  void dispose() {
    _nameCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: const Text('创建场景'),
        actions: [
          TextButton(
            onPressed: _saveScene,
            child: const Text('保存', style: TextStyle(color: Color(0xFF007DFF), fontSize: 16)),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _buildBasicInfo(),
          const SizedBox(height: 12),
          _buildTypeSelector(),
          if (_type == 'auto') ...[
            const SizedBox(height: 12),
            _buildConditionsSection(),
          ],
          const SizedBox(height: 12),
          _buildActionsSection(),
        ],
      ),
    );
  }

  Widget _buildBasicInfo() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('基本信息', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 16),
            TextField(
              controller: _nameCtrl,
              decoration: const InputDecoration(
                labelText: '场景名称',
                hintText: '例如：回家模式',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            const Text('选择图标', style: TextStyle(fontSize: 14, color: Color(0xFF666666))),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _icons.map((icon) => GestureDetector(
                onTap: () => setState(() => _selectedIcon = icon),
                child: Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: _selectedIcon == icon
                        ? const Color(0xFF007DFF).withOpacity(0.1)
                        : const Color(0xFFF5F5F5),
                    borderRadius: BorderRadius.circular(12),
                    border: _selectedIcon == icon
                        ? Border.all(color: const Color(0xFF007DFF), width: 2)
                        : null,
                  ),
                  child: Icon(_iconData(icon), size: 24, color: const Color(0xFF007DFF)),
                ),
              )).toList(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTypeSelector() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('场景类型', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(child: _typeOption('manual', '手动场景', '点击执行')),
                const SizedBox(width: 12),
                Expanded(child: _typeOption('auto', '自动化', '条件触发')),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _typeOption(String type, String title, String subtitle) {
    final selected = _type == type;
    return GestureDetector(
      onTap: () => setState(() => _type = type),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFF007DFF).withOpacity(0.05) : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: selected ? const Color(0xFF007DFF) : const Color(0xFFEEEEEE),
            width: selected ? 2 : 1,
          ),
        ),
        child: Column(
          children: [
            Icon(
              type == 'manual' ? Icons.touch_app : Icons.auto_awesome,
              color: selected ? const Color(0xFF007DFF) : const Color(0xFF999999),
              size: 28,
            ),
            const SizedBox(height: 8),
            Text(title, style: TextStyle(
              fontWeight: FontWeight.w500,
              color: selected ? const Color(0xFF007DFF) : const Color(0xFF333333),
            )),
            const SizedBox(height: 4),
            Text(subtitle, style: const TextStyle(fontSize: 11, color: Color(0xFF999999))),
          ],
        ),
      ),
    );
  }

  Widget _buildConditionsSection() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('触发条件', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                TextButton.icon(
                  onPressed: _addCondition,
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('添加'),
                ),
              ],
            ),
            if (_conditions.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 20),
                child: Center(
                  child: Text('暂无条件，点击"添加"创建', style: TextStyle(color: Color(0xFF999999))),
                ),
              )
            else
              ..._conditions.asMap().entries.map((entry) {
                final i = entry.key;
                final c = entry.value;
                return Container(
                  margin: const EdgeInsets.only(bottom: 8),
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8F8F8),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          c.type == 'time'
                              ? '时间: ${c.time ?? "未设置"}'
                              : '设备: ${c.deviceName ?? "未选择"} ${c.operator ?? ""} ${c.value ?? ""}',
                          style: const TextStyle(fontSize: 13),
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close, size: 18, color: Color(0xFF999999)),
                        onPressed: () => setState(() => _conditions.removeAt(i)),
                      ),
                    ],
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }

  Widget _buildActionsSection() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('执行动作', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                TextButton.icon(
                  onPressed: _addAction,
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('添加'),
                ),
              ],
            ),
            if (_actions.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 20),
                child: Center(
                  child: Text('暂无动作，点击"添加"创建', style: TextStyle(color: Color(0xFF999999))),
                ),
              )
            else
              ..._actions.asMap().entries.map((entry) {
                final i = entry.key;
                final a = entry.value;
                return Container(
                  margin: const EdgeInsets.only(bottom: 8),
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8F8F8),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.play_arrow, color: Color(0xFF007DFF), size: 20),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          '${a.deviceName ?? "未选择"} → ${a.property ?? ""} = ${a.value ?? ""}',
                          style: const TextStyle(fontSize: 13),
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close, size: 18, color: Color(0xFF999999)),
                        onPressed: () => setState(() => _actions.removeAt(i)),
                      ),
                    ],
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }

  void _addCondition() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text('选择条件类型', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            ),
            ListTile(
              leading: const Icon(Icons.access_time, color: Color(0xFF007DFF)),
              title: const Text('时间条件'),
              subtitle: const Text('到达指定时间时触发'),
              onTap: () {
                Navigator.pop(context);
                _addTimeCondition();
              },
            ),
            ListTile(
              leading: const Icon(Icons.devices_other, color: Color(0xFF4CAF50)),
              title: const Text('设备条件'),
              subtitle: const Text('设备属性满足条件时触发'),
              onTap: () {
                Navigator.pop(context);
                _addDeviceCondition();
              },
            ),
          ],
        ),
      ),
    );
  }

  void _addTimeCondition() {
    final timeCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('设置时间'),
        content: TextField(
          controller: timeCtrl,
          decoration: const InputDecoration(
            hintText: '例如: 08:00',
            border: OutlineInputBorder(),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () {
              setState(() {
                _conditions.add(SceneCondition(type: 'time', time: timeCtrl.text));
              });
              Navigator.pop(ctx);
            },
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _addDeviceCondition() {
    final deviceCtrl = TextEditingController();
    final propCtrl = TextEditingController();
    final valueCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('设备条件'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: deviceCtrl, decoration: const InputDecoration(labelText: '设备名称')),
            const SizedBox(height: 12),
            TextField(controller: propCtrl, decoration: const InputDecoration(labelText: '属性')),
            const SizedBox(height: 12),
            TextField(controller: valueCtrl, decoration: const InputDecoration(labelText: '值')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () {
              setState(() {
                _conditions.add(SceneCondition(
                  type: 'device',
                  deviceName: deviceCtrl.text,
                  property: propCtrl.text,
                  operator: '=',
                  value: valueCtrl.text,
                ));
              });
              Navigator.pop(ctx);
            },
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _addAction() {
    final deviceCtrl = TextEditingController();
    final propCtrl = TextEditingController();
    final valueCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('添加动作'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: deviceCtrl, decoration: const InputDecoration(labelText: '设备名称')),
            const SizedBox(height: 12),
            TextField(controller: propCtrl, decoration: const InputDecoration(labelText: '属性')),
            const SizedBox(height: 12),
            TextField(controller: valueCtrl, decoration: const InputDecoration(labelText: '值')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          TextButton(
            onPressed: () {
              setState(() {
                _actions.add(SceneAction(
                  deviceName: deviceCtrl.text,
                  property: propCtrl.text,
                  value: valueCtrl.text,
                ));
              });
              Navigator.pop(ctx);
            },
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  Future<void> _saveScene() async {
    if (_nameCtrl.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请输入场景名称')));
      return;
    }
    if (_actions.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请至少添加一个执行动作')));
      return;
    }

    final body = {
      'name': _nameCtrl.text,
      'icon': _selectedIcon,
      'type': _type,
      'conditions': _conditions.map((c) => {
        'type': c.type,
        if (c.deviceId != null) 'device_id': c.deviceId,
        if (c.deviceName != null) 'device_name': c.deviceName,
        if (c.property != null) 'property': c.property,
        if (c.operator != null) 'operator': c.operator,
        if (c.value != null) 'value': c.value,
        if (c.time != null) 'time': c.time,
      }).toList(),
      'actions': _actions.map((a) => {
        'type': a.type,
        if (a.deviceId != null) 'device_id': a.deviceId,
        if (a.deviceName != null) 'device_name': a.deviceName,
        if (a.property != null) 'property': a.property,
        if (a.value != null) 'value': a.value,
      }).toList(),
    };

    final provider = context.read<SceneProvider>();
    final ok = await provider.createScene(body);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(ok ? '场景创建成功' : '创建失败')),
      );
      if (ok) Navigator.pop(context);
    }
  }

  IconData _iconData(String iconName) {
    switch (iconName) {
      case 'home': return Icons.home;
      case 'exit_to_app': return Icons.exit_to_app;
      case 'bedtime': return Icons.bedtime;
      case 'alarm': return Icons.alarm;
      case 'lightbulb': return Icons.lightbulb;
      case 'security': return Icons.security;
      case 'temperature': return Icons.thermostat;
      case 'curtains': return Icons.curtains;
      default: return Icons.play_circle;
    }
  }
}
