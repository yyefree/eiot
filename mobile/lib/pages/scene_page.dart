import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/scene_provider.dart';
import '../models/models.dart';
import 'scene_create_page.dart';

class ScenePage extends StatefulWidget {
  const ScenePage({super.key});

  @override
  State<ScenePage> createState() => _ScenePageState();
}

class _ScenePageState extends State<ScenePage> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final sceneProvider = context.watch<SceneProvider>();

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: const Text('智能场景'),
        bottom: TabBar(
          controller: _tabController,
          labelColor: const Color(0xFF007DFF),
          unselectedLabelColor: const Color(0xFF999999),
          indicatorColor: const Color(0xFF007DFF),
          tabs: const [
            Tab(text: '手动场景'),
            Tab(text: '自动化'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildSceneList(sceneProvider.manualScenes, sceneProvider, false),
          _buildSceneList(sceneProvider.autoScenes, sceneProvider, true),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          await Navigator.push(context, MaterialPageRoute(builder: (_) => const SceneCreatePage()));
          sceneProvider.loadScenes();
        },
        backgroundColor: const Color(0xFF007DFF),
        child: const Icon(Icons.add, color: Colors.white),
      ),
    );
  }

  Widget _buildSceneList(List<Scene> scenes, SceneProvider provider, bool isAuto) {
    if (provider.loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (scenes.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              isAuto ? Icons.auto_awesome : Icons.play_circle_outline,
              size: 80,
              color: const Color(0xFFDDDDDD),
            ),
            const SizedBox(height: 16),
            Text(
              isAuto ? '暂无自动化规则' : '暂无手动场景',
              style: const TextStyle(fontSize: 16, color: Color(0xFF999999)),
            ),
            const SizedBox(height: 8),
            const Text(
              '点击右下角 + 创建',
              style: TextStyle(fontSize: 13, color: Color(0xFFCCCCCC)),
            ),
          ],
        ),
      );
    }
    return RefreshIndicator(
      onRefresh: provider.loadScenes,
      child: ListView.builder(
        padding: const EdgeInsets.all(12),
        itemCount: scenes.length,
        itemBuilder: (context, index) => _buildSceneCard(scenes[index], provider),
      ),
    );
  }

  Widget _buildSceneCard(Scene scene, SceneProvider provider) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    color: const Color(0xFF007DFF).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    _sceneIcon(scene.icon),
                    color: const Color(0xFF007DFF),
                    size: 24,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        scene.name,
                        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
                      ),
                      if (scene.type == 'auto' && scene.conditions.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 4),
                          child: Text(
                            _buildConditionText(scene.conditions.first),
                            style: const TextStyle(fontSize: 12, color: Color(0xFF999999)),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                    ],
                  ),
                ),
                if (scene.type == 'auto')
                  Switch(
                    value: scene.enabled,
                    onChanged: (v) => provider.toggleScene(scene.id, v),
                    activeColor: const Color(0xFF007DFF),
                  )
                else
                  TextButton(
                    onPressed: () => provider.runScene(scene.id),
                    style: TextButton.styleFrom(
                      backgroundColor: const Color(0xFF007DFF),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                    child: const Text('执行', style: TextStyle(fontSize: 13)),
                  ),
              ],
            ),
            if (scene.actions.isNotEmpty) ...[
              const Divider(height: 24),
              Wrap(
                spacing: 8,
                runSpacing: 4,
                children: scene.actions.take(3).map((a) => Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF5F5F5),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    '${a.deviceName ?? ''} → ${a.property ?? ''} = ${a.value ?? ''}',
                    style: const TextStyle(fontSize: 11, color: Color(0xFF666666)),
                  ),
                )).toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }

  IconData _sceneIcon(String iconName) {
    switch (iconName) {
      case 'home': return Icons.home;
      case 'exit': return Icons.exit_to_app;
      case 'bedtime': return Icons.bedtime;
      case 'alarm': return Icons.alarm;
      case 'light': return Icons.lightbulb;
      case 'security': return Icons.security;
      default: return Icons.play_circle;
    }
  }

  String _buildConditionText(SceneCondition condition) {
    if (condition.type == 'time') {
      return '当时间到达 ${condition.time ?? ''}';
    }
    return '当 ${condition.deviceName ?? ''} ${condition.operator ?? ''} ${condition.value ?? ''}';
  }
}
