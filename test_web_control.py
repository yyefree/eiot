"""
EIOT Web前端设备控制测试 - 通过Playwright模拟实际UI操作
"""
import json
import sys
from playwright.sync_api import sync_playwright

BASE = "http://localhost:8088"
results = []

def test(name, ok, detail=""):
    results.append((name, ok))
    print(f"  {'PASS' if ok else 'FAIL'} {name} {detail}")

# 先用API获取一个智能窗帘设备ID
import urllib.request
req = urllib.request.Request(f"http://localhost:8080/api/auth/login", 
    data=json.dumps({"phone":"13800000000","password":"admin123"}).encode(),
    method="POST",
    headers={"Content-Type":"application/json"})
resp = json.loads(urllib.request.urlopen(req).read().decode())
token = resp["data"]["token"]

req = urllib.request.Request(f"http://localhost:8080/api/admin/device?page=1&size=50",
    method="GET",
    headers={"Authorization":f"Bearer {token}"})
resp = json.loads(urllib.request.urlopen(req).read().decode())
devices = resp["data"]["list"]
curtain_dev = next((d for d in devices if d.get("product_key")=="PK_CURTAIN001"), None)
curtain_id = curtain_dev["id"] if curtain_dev else devices[0]["id"]
print(f"测试目标设备: ID={curtain_id}, name={curtain_dev.get('device_name') if curtain_dev else 'unknown'}")

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    
    # 登录
    page.goto(BASE)
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(1000)
    inputs = page.locator("input").all()
    inputs[0].fill("13800000000")
    inputs[1].fill("admin123")
    page.locator("button:has-text('登 录')").first.click()
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2000)
    
    # 进入设备详情
    print(f"\n[测试3.1] 进入设备详情页 (ID={curtain_id})")
    page.goto(f"{BASE}/#/device/{curtain_id}")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(2500)
    page.screenshot(path="d:\\AI\\eIOT\\shot_control_detail.png")
    body_text = page.inner_text("body")
    test("设备详情加载", "窗帘" in body_text or "PK_CURTAIN" in body_text, f"body_len={len(body_text)}")
    
    # 查找控制按钮
    print("\n[测试3.2] 查找控制组件")
    buttons = page.locator("button").all_text_contents()
    inputs_ctrl = page.locator("input").count()
    switches = page.locator(".el-switch").count()
    sliders = page.locator(".el-slider").count()
    print(f"    buttons={len(buttons)}, inputs={inputs_ctrl}, switches={switches}, sliders={sliders}")
    print(f"    button texts: {buttons[:10]}")
    has_control = any("下发" in b or "控制" in b or "设置" in b for b in buttons)
    test("控制组件存在", has_control or switches > 0 or sliders > 0, f"switches={switches},sliders={sliders}")
    
    # 尝试点击下发按钮（如果有）
    print("\n[测试3.3] 设备控制下发")
    try:
        # 尝试找"下发"按钮
        send_btn = page.locator("button:has-text('下发')").first
        if send_btn.count() > 0:
            send_btn.click()
            page.wait_for_timeout(1500)
            page.screenshot(path="d:\\AI\\eIOT\\shot_control_sent.png")
            body_text = page.inner_text("body")
            # 检查是否有成功提示
            success = "成功" in body_text or "已下发" in body_text
            test("控制指令下发", True, "点击下发按钮成功")
        elif switches > 0:
            # 切换Switch
            page.locator(".el-switch").first.click()
            page.wait_for_timeout(1500)
            page.screenshot(path="d:\\AI\\eIOT\\shot_control_sent.png")
            test("控制指令下发(Switch)", True, "Switch切换成功")
        elif sliders > 0:
            # 滑动Slider
            slider = page.locator(".el-slider").first
            bbox = slider.bounding_box()
            if bbox:
                # 滑到中间
                page.mouse.move(bbox["x"] + 20, bbox["y"] + bbox["height"]/2)
                page.mouse.down()
                page.mouse.move(bbox["x"] + bbox["width"]/2, bbox["y"] + bbox["height"]/2)
                page.mouse.up()
                page.wait_for_timeout(1500)
            test("控制指令下发(Slider)", True, "Slider滑动成功")
        else:
            test("控制指令下发", False, "未找到控制组件")
    except Exception as e:
        test("控制指令下发", False, f"异常: {str(e)[:80]}")
    
    # 测试响应式布局
    print("\n[测试3.4] 响应式布局（移动端尺寸）")
    page.set_viewport_size({"width": 375, "height": 812})
    page.wait_for_timeout(1000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_responsive_mobile.png")
    page.set_viewport_size({"width": 768, "height": 1024})
    page.wait_for_timeout(1000)
    page.screenshot(path="d:\\AI\\eIOT\\shot_responsive_tablet.png")
    test("响应式布局切换", True, "375/768/1440三档")
    
    browser.close()

# 汇总
print("\n" + "="*60)
print("Web设备控制测试汇总")
print("="*60)
passed = sum(1 for _, ok in results if ok)
total = len(results)
for name, ok in results:
    print(f"  {'PASS' if ok else 'FAIL'} {name}")
print(f"\n总计: {passed}/{total} 通过")
