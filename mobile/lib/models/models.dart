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
      id: json['id'] as int? ?? 0,
      username: json['username'] as String? ?? '',
      phone: json['phone'] as String? ?? '',
      nickname: json['nickname'] as String? ?? '',
      role: json['role'] as String? ?? 'user',
      status: json['status'] as int? ?? 1,
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

  DeviceItem({
    required this.id,
    this.name = '',
    this.deviceSn = '',
    this.productId = 0,
    this.productKey = '',
    this.productName = '',
    this.online = false,
    this.createdAt = '',
  });

  factory DeviceItem.fromJson(Map<String, dynamic> json) {
    return DeviceItem(
      id: json['id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      deviceSn: json['device_sn'] as String? ?? '',
      productId: json['product_id'] as int? ?? 0,
      productKey: json['product_key'] as String? ?? '',
      productName: json['productName'] as String? ?? '',
      online: json['online'] as bool? ?? false,
      createdAt: json['createdAt'] as String? ?? '',
    );
  }
}

class DeviceDetail {
  final DeviceItem device;
  final Map<String, String> latest;
  final List<ThingProperty> properties;

  DeviceDetail({
    required this.device,
    this.latest = const {},
    this.properties = const [],
  });
}

class ThingProperty {
  final String identifier;
  final String name;
  final String accessMode;
  final String dataType;
  final double? min;
  final double? max;
  final String unit;

  ThingProperty({
    required this.identifier,
    this.name = '',
    this.accessMode = 'rw',
    this.dataType = 'string',
    this.min,
    this.max,
    this.unit = '',
  });

  static double? _toDouble(dynamic v) {
    if (v == null) return null;
    if (v is num) return v.toDouble();
    return double.tryParse(v.toString());
  }

  factory ThingProperty.fromJson(Map<String, dynamic> json) {
    final dt = json['dataType'];
    String typeStr = 'string';
    Map<String, dynamic>? specs;
    if (dt is Map) {
      typeStr = dt['type'] as String? ?? 'string';
      specs = dt['specs'] as Map<String, dynamic>?;
    } else if (dt is String) {
      typeStr = dt;
    }
    return ThingProperty(
      identifier: json['identifier'] as String? ?? '',
      name: json['name'] as String? ?? '',
      accessMode: json['accessMode'] as String? ?? 'rw',
      dataType: typeStr,
      min: _toDouble(specs?['min']),
      max: _toDouble(specs?['max']),
      unit: specs?['unit'] as String? ?? '',
    );
  }

  bool get canWrite => accessMode.contains('w');
  bool get isBool => dataType == 'bool';
  bool get isNumber => dataType == 'int' || dataType == 'float' || dataType == 'double';
}
