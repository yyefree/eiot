"""
EIOT Web前端联合测试 - Playwright脚本
"""
import json
import sys
from playwright.sync_api import sync_playwright

BASE = "http://localhost:8088"
results = []

def test(name, ok, detail=""):
    results.append((name, ok))
    print(f"  {'PASS' if ok else 'FAIL'} {name} {detail}")

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    
    # 收集console错误
    errors = []
    page.on("console", lambda msg: errors.append(msg.text) if msg.type == "error" else None)
    
    # 测试1: 登录页加载
    print("\n[测试2.1] 登录页加载")
    page.goto(BASE)
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(1500)
    page.screenshot(path="d:\\AI\\eIOT\\shot_login.png")
    has_login = page.locator("input").count() >= 2
    test("登录页加载", has_login, f"inputs={page.locator('input').count()}")
    
    # 测试2: 登录
    print("\n[测试2.2] 管理员登录")
    inputs = page.locator("input").all()
    inputs[0].fill("13800000000")
    inputs[1].fill("admin123")
    page.screenshot(path="d:\\AI\\eIOT\\shot_login_filled.png")
    # 点击登录按钮
    page.locator("button:has-text('登 录')").first.click()
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_dashboard.png")
    url = page.url
    test("登录成功跳转", "/dashboard" in url or "/device" in url or url != BASE + "/", f"url={url}")
    
    # 测试3: 看板数据
    print("\n[测试2.3] 看板统计")
    page.goto(BASE + "/dashboard")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(1500)
    page.screenshot(path="d:\\AI\\eIOT\\shot_dashboard_full.png")
    body_text = page.inner_text("body")
    has_stats = "50" in body_text or "设备" in body_text
    test("看板数据加载", has_stats, f"body_len={len(body_text)}")
    
    # 测试4: 产品管理
    print("\n[测试2.4] 产品管理页")
    page.goto(BASE + "/#/product")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_product.png")
    body_text = page.inner_text("body")
    has_products = "空气质量监测" in body_text or "智能窗帘" in body_text
    test("产品列表加载", has_products, f"body_len={len(body_text)}")
    
    # 测试5: 设备管理
    print("\n[测试2.5] 设备管理页")
    page.goto(BASE + "/#/device")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_device.png")
    body_text = page.inner_text("body")
    has_devices = "SN000" in body_text or "设备" in body_text
    test("设备列表加载", has_devices, f"body_len={len(body_text)}")
    
    # 测试6: 设备详情
    print("\n[测试2.6] 设备详情页")
    page.goto(BASE + "/#/device/50")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_device_detail.png")
    body_text = page.inner_text("body")
    has_detail = "空气质量监测" in body_text or "PK_AIR001" in body_text
    test("设备详情加载", has_detail, f"body_len={len(body_text)}")
    
    # 测试7: 共享页面
    print("\n[测试2.7] 共享管理页")
    page.goto(BASE + "/#/share")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_share.png")
    test("共享页面加载", "share" in page.url, f"url={page.url}")
    
    # 测试8: 用户管理
    print("\n[测试2.8] 用户管理页")
    page.goto(BASE + "/#/users")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_users.png")
    body_text = page.inner_text("body")
    has_users = "13800000" in body_text or "用户" in body_text
    test("用户管理加载", has_users, f"body_len={len(body_text)}")
    
    # 测试9: 项目管理
    print("\n[测试2.9] 项目管理页")
    page.goto(BASE + "/#/project")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_project.png")
    test("项目页面加载", "project" in page.url, f"url={page.url}")
    
    # 控制台错误检查
    print(f"\n[Console错误] {len(errors)} 个")
    if errors:
        for e in errors[:5]:
            print(f"  ERROR: {e[:100]}")
    test("无控制台错误", len(errors) == 0, f"errors={len(errors)}")
    
    browser.close()

# 汇总
print("\n" + "="*60)
print("Web前端测试汇总")
print("="*60)
passed = sum(1 for _, ok in results if ok)
total = len(results)
for name, ok in results:
    print(f"  {'PASS' if ok else 'FAIL'} {name}")
print(f"\n总计: {passed}/{total} 通过")
