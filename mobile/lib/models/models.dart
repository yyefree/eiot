// Safe JSON parsing utilities
class JsonUtils {
  static String getString(dynamic v, [String def = '']) {
    if (v == null) return def;
    if (v is String) return v;
    return v.toString();
  }

  static int getInt(dynamic v, [int def = 0]) {
    if (v == null) return def;
    if (v is int) return v;
    if (v is double) return v.toInt();
    if (v is String) return int.tryParse(v) ?? def;
    return def;
  }

  static double getDouble(dynamic v, [double def = 0.0]) {
    if (v == null) return def;
    if (v is double) return v;
    if (v is int) return v.toDouble();
    if (v is String) return double.tryParse(v) ?? def;
    return def;
  }

  static bool getBool(dynamic v, [bool def = false]) {
    if (v == null) return def;
    if (v is bool) return v;
    if (v is String) return v.toLowerCase() == 'true' || v == '1';
    if (v is int) return v != 0;
    return def;
  }

  static List<String> getStringList(dynamic v) {
    if (v == null) return [];
    if (v is List) {
      return v.map((e) => e?.toString() ?? '').where((s) => s.isNotEmpty).toList();
    }
    return [];
  }

  static Map<String, dynamic> getMap(dynamic v) {
    if (v == null) return {};
    if (v is Map<String, dynamic>) return v;
    if (v is Map) return Map<String, dynamic>.from(v);
    return {};
  }
}
class UserInfo {
  final int id;
  final String username;
  final String phone;
  final String nickname;
  final String role;
  final int status;

  UserInfo({
    required this.id,
    this.username = '',
    this.phone = '',
    this.nickname = '',
    this.role = 'user',
    this.status = 1,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: JsonUtils.getInt(json['id']),
      username: JsonUtils.getString(json['username']),
      phone: JsonUtils.getString(json['phone']),
      nickname: JsonUtils.getString(json['nickname']),
      role: JsonUtils.getString(json['role'], 'user'),
      status: JsonUtils.getInt(json['status'], 1),
    );
  }

  bool get isAdmin => role == 'admin';
}

class DeviceItem {
  final int id;
  final String name;
  final String deviceSn;
  final int productId;
  final String productKey;
  final String productName;
  final bool online;
  final String createdAt;
  final int? roomId;

  DeviceItem({
    required this.id,
    this.name = '',
    this.deviceSn = '',
    this.productId = 0,
    this.productKey = '',
    this.productName = '',
    this.online = false,
    this.createdAt = '',
    this.roomId,
  });

  factory DeviceItem.fromJson(Map<String, dynamic> json) {
    return DeviceItem(
      id: JsonUtils.getInt(json['id']),
      name: JsonUtils.getString(json['name']),
      deviceSn: JsonUtils.getString(json['device_sn']),
      productId: JsonUtils.getInt(json['product_id']),
      productKey: JsonUtils.getString(json['product_key']),
      productName: JsonUtils.getString(json['productName']),
      online: JsonUtils.getBool(json['online']),
      createdAt: JsonUtils.getString(json['createdAt']),
      roomId: json['room_id'] == null ? null : JsonUtils.getInt(json['room_id']),
    );
  }
}

class DeviceDetail {
  final DeviceItem device;
  final Map<String, String> latest;
  final List<ThingProperty> properties;
  final List<ThingEvent> events;
  final List<ThingService> services;

  DeviceDetail({
    required this.device,
    this.latest = const {},
    this.properties = const [],
    this.events = const [],
    this.services = const [],
  });

  factory DeviceDetail.fromJson(Map<String, dynamic> json) {
    final props = <ThingProperty>[];
    final tm = json['thingModel'];
    if (tm is Map && tm['properties'] is List) {
      for (var p in tm['properties']) {
        if (p is Map<String, dynamic>) props.add(ThingProperty.fromJson(p));
        else if (p is Map) props.add(ThingProperty.fromJson(Map<String, dynamic>.from(p)));
      }
    }
    final events = <ThingEvent>[];
    if (tm is Map && tm['events'] is List) {
      for (var e in tm['events']) {
        if (e is Map<String, dynamic>) events.add(ThingEvent.fromJson(e));
        else if (e is Map) events.add(ThingEvent.fromJson(Map<String, dynamic>.from(e)));
      }
    }
    final services = <ThingService>[];
    if (tm is Map && tm['services'] is List) {
      for (var s in tm['services']) {
        if (s is Map<String, dynamic>) services.add(ThingService.fromJson(s));
        else if (s is Map) services.add(ThingService.fromJson(Map<String, dynamic>.from(s)));
      }
    }
    return DeviceDetail(
      device: DeviceItem.fromJson((json['device'] as Map?)?.cast<String, dynamic>() ?? {}),
      latest: (json['latest'] as Map?)?.cast<String, String>() ?? {},
      properties: props,
      events: events,
      services: services,
    );
  }
}

class ThingProperty {
  final String identifier;
  final String name;
  final String accessMode;
  final String dataType;
  final double? min;
  final double? max;
  final String unit;
  final String description;
  final bool required;
  final double? step;
  final List<String>? enumValues;

  ThingProperty({
    required this.identifier,
    this.name = '',
    this.accessMode = 'rw',
    this.dataType = 'string',
    this.min,
    this.max,
    this.unit = '',
    this.description = '',
    this.required = false,
    this.step,
    this.enumValues,
  });

  factory ThingProperty.fromJson(Map<String, dynamic> json) {
    final dt = json['dataType'];
    String typeStr = 'string';
    Map<String, dynamic>? specs;
    if (dt is Map) {
      typeStr = JsonUtils.getString(dt['type'], 'string');
      specs = dt['specs'] as Map<String, dynamic>?;
    } else if (dt is String) {
      typeStr = dt;
    }
    List<String>? enums;
    if (specs?['enumValues'] is List) {
      enums = JsonUtils.getStringList(specs!['enumValues']);
    } else if (json['enumValues'] is List) {
      enums = JsonUtils.getStringList(json['enumValues']);
    }
    return ThingProperty(
      identifier: JsonUtils.getString(json['identifier']),
      name: JsonUtils.getString(json['name']),
      accessMode: JsonUtils.getString(json['accessMode'], 'rw'),
      dataType: typeStr,
      min: JsonUtils.getDouble(specs?['min']),
      max: JsonUtils.getDouble(specs?['max']),
      unit: JsonUtils.getString(specs?['unit']),
      description: JsonUtils.getString(json['description']),
      required: JsonUtils.getBool(json['required']),
      step: JsonUtils.getDouble(specs?['step']),
      enumValues: enums,
    );
  }

  bool get canWrite => accessMode.contains('w');
  bool get isBool => dataType == 'bool';
  bool get isNumber => dataType == 'int' || dataType == 'float' || dataType == 'double';
  bool get isEnum => enumValues != null && enumValues!.isNotEmpty;
}

class ThingEvent {
  final String identifier;
  final String name;
  final String type;
  final String description;
  final List<EventOutputParam> outputParams;

  ThingEvent({
    required this.identifier,
    required this.name,
    this.type = 'info',
    this.description = '',
    this.outputParams = const [],
  });

  factory ThingEvent.fromJson(Map<String, dynamic> json) {
    final output = <EventOutputParam>[];
    if (json['outputParams'] is List) {
      for (var p in json['outputParams']) {
        output.add(EventOutputParam.fromJson(p));
      }
    }
    return ThingEvent(
      identifier: json['identifier']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      type: json['type']?.toString() ?? 'info',
      description: json['description']?.toString() ?? '',
      outputParams: output,
    );
  }
}

class EventOutputParam {
  final String identifier;
  final String name;
  final String dataType;
  final String unit;

  EventOutputParam({
    required this.identifier,
    required this.name,
    this.dataType = 'string',
    this.unit = '',
  });

  factory EventOutputParam.fromJson(Map<String, dynamic> json) {
    return EventOutputParam(
      identifier: json['identifier']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      dataType: json['dataType']?.toString() ?? 'string',
      unit: json['unit']?.toString() ?? '',
    );
  }
}

class ThingService {
  final String identifier;
  final String name;
  final String description;
  final String callType;
  final List<ServiceParam> inputParams;
  final List<ServiceParam> outputParams;

  ThingService({
    required this.identifier,
    required this.name,
    this.description = '',
    this.callType = 'sync',
    this.inputParams = const [],
    this.outputParams = const [],
  });

  factory ThingService.fromJson(Map<String, dynamic> json) {
    final inputs = <ServiceParam>[];
    if (json['inputParams'] is List) {
      for (var p in json['inputParams']) {
        inputs.add(ServiceParam.fromJson(p));
      }
    }
    final outputs = <ServiceParam>[];
    if (json['outputParams'] is List) {
      for (var p in json['outputParams']) {
        outputs.add(ServiceParam.fromJson(p));
      }
    }
    return ThingService(
      identifier: json['identifier']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      callType: json['callType']?.toString() ?? 'sync',
      inputParams: inputs,
      outputParams: outputs,
    );
  }
}

class ServiceParam {
  final String identifier;
  final String name;
  final String dataType;
  final String unit;

  ServiceParam({
    required this.identifier,
    required this.name,
    this.dataType = 'string',
    this.unit = '',
  });

  factory ServiceParam.fromJson(Map<String, dynamic> json) {
    return ServiceParam(
      identifier: json['identifier']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      dataType: json['dataType']?.toString() ?? 'string',
      unit: json['unit']?.toString() ?? '',
    );
  }
}

class Home {
  final int id;
  final String name;
  final String address;
  final int memberCount;
  final bool isDefault;
  final String createdAt;

  Home({
    required this.id,
    this.name = '',
    this.address = '',
    this.memberCount = 1,
    this.isDefault = false,
    this.createdAt = '',
  });

  factory Home.fromJson(Map<String, dynamic> json) {
    return Home(
      id: json['id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      address: json['address'] as String? ?? '',
      memberCount: json['member_count'] as int? ?? 1,
      isDefault: json['is_default'] as bool? ?? false,
      createdAt: json['createdAt'] as String? ?? '',
    );
  }
}

class HomeMember {
  final int id;
  final int userId;
  final String nickname;
  final String phone;
  final String role;
  final String joinedAt;

  HomeMember({
    required this.id,
    this.userId = 0,
    this.nickname = '',
    this.phone = '',
    this.role = 'member',
    this.joinedAt = '',
  });

  factory HomeMember.fromJson(Map<String, dynamic> json) {
    return HomeMember(
      id: json['id'] as int? ?? 0,
      userId: json['user_id'] as int? ?? 0,
      nickname: json['nickname'] as String? ?? '',
      phone: json['phone'] as String? ?? '',
      role: json['role'] as String? ?? 'member',
      joinedAt: json['joinedAt'] as String? ?? '',
    );
  }
}

class Room {
  final int id;
  final String name;
  final String icon;
  final int deviceCount;
  final int sortOrder;

  Room({
    required this.id,
    this.name = '',
    this.icon = 'meeting_room',
    this.deviceCount = 0,
    this.sortOrder = 0,
  });

  factory Room.fromJson(Map<String, dynamic> json) {
    return Room(
      id: json['id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      icon: json['icon'] as String? ?? 'meeting_room',
      deviceCount: json['device_count'] as int? ?? 0,
      sortOrder: json['sort_order'] as int? ?? 0,
    );
  }
}

class Scene {
  final int id;
  final String name;
  final String icon;
  final String type;
  final bool enabled;
  final List<SceneCondition> conditions;
  final List<SceneAction> actions;
  final String createdAt;

  Scene({
    required this.id,
    this.name = '',
    this.icon = 'play_circle',
    this.type = 'manual',
    this.enabled = true,
    this.conditions = const [],
    this.actions = const [],
    this.createdAt = '',
  });

  factory Scene.fromJson(Map<String, dynamic> json) {
    return Scene(
      id: json['id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      icon: json['icon'] as String? ?? 'play_circle',
      type: json['type'] as String? ?? 'manual',
      enabled: json['enabled'] as bool? ?? true,
      conditions: (json['conditions'] as List?)
              ?.whereType<Map<String, dynamic>>()
              .map((e) => SceneCondition.fromJson(e))
              .toList() ??
          [],
      actions: (json['actions'] as List?)
              ?.whereType<Map<String, dynamic>>()
              .map((e) => SceneAction.fromJson(e))
              .toList() ??
          [],
      createdAt: json['createdAt'] as String? ?? '',
    );
  }
}

class SceneCondition {
  final String type;
  final String? deviceId;
  final String? deviceName;
  final String? property;
  final String? operator;
  final dynamic value;
  final String? time;
  final String? days;

  SceneCondition({
    this.type = 'device',
    this.deviceId,
    this.deviceName,
    this.property,
    this.operator,
    this.value,
    this.time,
    this.days,
  });

  factory SceneCondition.fromJson(Map<String, dynamic> json) {
    return SceneCondition(
      type: json['type'] as String? ?? 'device',
      deviceId: json['device_id']?.toString(),
      deviceName: json['device_name'] as String?,
      property: json['property'] as String?,
      operator: json['operator'] as String?,
      value: json['value'],
      time: json['time'] as String?,
      days: json['days'] as String?,
    );
  }
}

class SceneAction {
  final String type;
  final String? deviceId;
  final String? deviceName;
  final String? property;
  final dynamic value;
  final String? sceneId;

  SceneAction({
    this.type = 'device',
    this.deviceId,
    this.deviceName,
    this.property,
    this.value,
    this.sceneId,
  });

  factory SceneAction.fromJson(Map<String, dynamic> json) {
    return SceneAction(
      type: json['type'] as String? ?? 'device',
      deviceId: json['device_id']?.toString(),
      deviceName: json['device_name'] as String?,
      property: json['property'] as String?,
      value: json['value'],
      sceneId: json['scene_id']?.toString(),
    );
  }
}

class Message {
  final int id;
  final String type;
  final String title;
  final String content;
  final bool read;
  final String createdAt;

  Message({
    required this.id,
    this.type = 'system',
    this.title = '',
    this.content = '',
    this.read = false,
    this.createdAt = '',
  });

  factory Message.fromJson(Map<String, dynamic> json) {
    return Message(
      id: json['id'] as int? ?? 0,
      type: json['type'] as String? ?? 'system',
      title: json['title'] as String? ?? '',
      content: json['content'] as String? ?? '',
      read: json['read'] as bool? ?? false,
      createdAt: json['createdAt'] as String? ?? '',
    );
  }
}

class OTAFirmware {
  final int id;
  final String version;
  final int productId;
  final String changelog;
  final int size;
  final String fileUrl;
  final String status;
  final String createdAt;

  OTAFirmware({
    required this.id,
    this.version = '',
    this.productId = 0,
    this.changelog = '',
    this.size = 0,
    this.fileUrl = '',
    this.status = '',
    this.createdAt = '',
  });

  factory OTAFirmware.fromJson(Map<String, dynamic> json) {
    return OTAFirmware(
      id: json['id'] as int? ?? 0,
      version: json['version'] as String? ?? '',
      productId: json['product_id'] as int? ?? 0,
      changelog: json['changelog'] as String? ?? '',
      size: json['size'] as int? ?? 0,
      fileUrl: json['file_url'] as String? ?? '',
      status: json['status'] as String? ?? '',
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}
