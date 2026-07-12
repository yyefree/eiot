"""
EIOT 三端联合测试 - API 调用脚本
"""
import json
import sys
import urllib.request
import urllib.error

BASE = "http://localhost:8080/api"

def call(method, path, token=None, body=None):
    url = f"{BASE}{path}"
    data = json.dumps(body).encode() if body else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", f"Bearer {token}")
    try:
        with urllib.request.urlopen(req, timeout=10) as r:
            return r.status, json.loads(r.read().decode())
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode()) if e.read() else {}

def main():
    results = []
    
    # 1. 密码登录
    print("\n[1] 密码登录 POST /api/auth/login")
    code, resp = call("POST", "/auth/login", body={"phone": "13800000000", "password": "admin123"})
    token = resp.get("data", {}).get("token") if code == 200 else None
    print(f"    Status: {code}, Token: {'OK' if token else 'FAIL'}")
    results.append(("密码登录", code == 200 and token is not None))
    
    if not token:
        print("登录失败，终止测试")
        return
    
    # 2. 用户信息
    print("\n[2] 用户信息 GET /api/user/info")
    code, resp = call("GET", "/user/info", token=token)
    user = resp.get("data", {})
    print(f"    Status: {code}, User: {user.get('username')}/{user.get('phone')}/role={user.get('role')}")
    results.append(("用户信息", code == 200 and user.get("username") == "admin"))
    
    # 3. 看板统计
    print("\n[3] 看板统计 GET /api/admin/stats")
    code, resp = call("GET", "/admin/stats", token=token)
    stats = resp.get("data", {})
    print(f"    Status: {code, }, Stats: products={stats.get('product_count')}, devices={stats.get('device_count')}, online={stats.get('online_count')}, users={stats.get('user_count')}")
    results.append(("看板统计", code == 200 and stats.get("product_count", 0) > 0))
    
    # 4. 产品列表
    print("\n[4] 产品列表 GET /api/admin/product")
    code, resp = call("GET", "/admin/product?page=1&size=20", token=token)
    products = resp.get("data", {}).get("list", []) if code == 200 else []
    print(f"    Status: {code}, Products: {len(products)}")
    for p in products[:3]:
        print(f"      - {p.get('name')} (key={p.get('product_key')})")
    results.append(("产品列表", code == 200 and len(products) > 0))
    
    # 5. 项目列表
    print("\n[5] 项目列表 GET /api/admin/project")
    code, resp = call("GET", "/admin/project?page=1&size=20", token=token)
    projects = resp.get("data", {}).get("list", []) if code == 200 else []
    print(f"    Status: {code, }, Projects: {len(projects)}")
    results.append(("项目列表", code == 200))
    
    # 6. 设备列表（管理员视角）
    print("\n[6] 设备列表 GET /api/admin/device")
    code, resp = call("GET", "/admin/device?page=1&size=20", token=token)
    devices_admin = resp.get("data", {}).get("list", []) if code == 200 else []
    print(f"    Status: {code, }, Devices: {len(devices_admin)}")
    results.append(("设备列表(admin)", code == 200 and len(devices_admin) > 0))
    
    # 7. 设备列表（用户视角）
    print("\n[7] 设备列表 GET /api/device")
    code, resp = call("GET", "/device?page=1&size=20", token=token)
    devices_user = resp.get("data", {}).get("list", []) if code == 200 else []
    print(f"    Status: {code, }, Devices: {len(devices_user)}")
    results.append(("设备列表(user)", code == 200 and len(devices_user) > 0))
    
    # 8. 设备详情
    if devices_user:
        dev_id = devices_user[0].get("id")
        print(f"\n[8] 设备详情 GET /api/device/{dev_id}")
        code, resp = call("GET", f"/device/{dev_id}", token=token)
        dev = resp.get("data", {}).get("device", {}) if code == 200 else {}
        latest = resp.get("data", {}).get("latest", {}) if code == 200 else {}
        tm = resp.get("data", {}).get("thingModel", {}) if code == 200 else {}
        props = tm.get("properties", []) if isinstance(tm, dict) else []
        print(f"    Status: {code, }, Device: {dev.get('device_name')}, Latest keys: {list(latest.keys())}, Props: {len(props)}")
        results.append(("设备详情", code == 200 and dev.get("device_name")))
    
    # 9. 设备控制下发
    if devices_user:
        dev_id = devices_user[0].get("id")
        # 找一个可写属性
        test_param = {}
        if devices_user[0].get("product_key") == "PK_CURTAIN001":
            test_param = {"open_percent": 50}
        elif devices_user[0].get("product_key") == "PK_AIR001":
            test_param = {"pm25": 35}
        else:
            test_param = {"test_value": 1}
        print(f"\n[9] 设备控制 POST /api/device/{dev_id}/control")
        code, resp = call("POST", f"/device/{dev_id}/control", token=token, body={"params": test_param})
        print(f"    Status: {code, }, Params: {test_param}")
        results.append(("设备控制", code == 200))
    
    # 10. 共享列表
    print("\n[10] 共享列表 GET /api/device/share")
    code, resp = call("GET", "/device/share?page=1&size=50", token=token)
    shares = resp.get("data", {}).get("list", []) if code == 200 else []
    print(f"    Status: {code, }, Shares: {len(shares)}")
    results.append(("共享列表", code == 200))
    
    # 11. 用户列表（管理员）
    print("\n[11] 用户列表 GET /api/admin/user")
    code, resp = call("GET", "/admin/user?page=1&size=20", token=token)
    users = resp.get("data", {}).get("list", []) if code == 200 else []
    print(f"    Status: {code, }, Users: {len(users)}")
    results.append(("用户列表", code == 200))
    
    # 12. 验证码发送
    print("\n[12] 验证码发送 POST /api/auth/send-code")
    code, resp = call("POST", "/auth/send-code", body={"phone": "13800000000"})
    vcode = resp.get("data", {}).get("code") if code == 200 else None
    print(f"    Status: {code, }, Code: {vcode}")
    results.append(("验证码发送", code == 200 and vcode is not None))
    
    # 13. 验证码登录
    if vcode:
        print("\n[13] 验证码登录 POST /api/auth/login-code")
        code, resp = call("POST", "/auth/login-code", body={"phone": "13800000000", "code": vcode})
        token2 = resp.get("data", {}).get("token") if code == 200 else None
        print(f"    Status: {code, }, Token: {'OK' if token2 else 'FAIL'}")
        results.append(("验证码登录", code == 200 and token2 is not None))
    
    # 汇总
    print("\n" + "="*60)
    print("API 测试汇总")
    print("="*60)
    passed = sum(1 for _, ok in results if ok)
    total = len(results)
    for name, ok in results:
        print(f"  {'PASS' if ok else 'FAIL'} {name}")
    print(f"\n总计: {passed}/{total} 通过")
    return passed == total

if __name__ == "__main__":
    main()
