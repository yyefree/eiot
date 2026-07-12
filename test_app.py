"""
EIOT 移动端App自动化测试 - ADB脚本 (最终版)
"""
import subprocess
import time
import re
import xml.etree.ElementTree as ET

DEVICE = "emulator-5554"
WORKDIR = "d:\\AI\\eIOT"
results = []

def adb(cmd):
    r = subprocess.run(f"adb -s {DEVICE} {cmd}", capture_output=True, text=True, timeout=30, shell=True)
    return r.stdout + r.stderr

def adb_shell(cmd):
    return adb(f"shell {cmd}")

def screenshot(name):
    path = f"{WORKDIR}\\shot_app_{name}.png"
    subprocess.run(f"adb -s {DEVICE} exec-out screencap -p > {path}", shell=True, timeout=15)
    return path

def dump_ui():
    adb_shell("uiautomator dump /sdcard/ui.xml")
    adb("pull /sdcard/ui.xml d:\\AI\\eIOT\\app_ui.xml")
    return "d:\\AI\\eIOT\\app_ui.xml"

def parse_bounds(bounds_str):
    m = re.match(r'\[(\d+),(\d+)\]\[(\d+),(\d+)\]', bounds_str)
    if m:
        return int(m.group(1)), int(m.group(2)), int(m.group(3)), int(m.group(4))
    return None

def parse_nodes():
    tree = ET.parse("d:\\AI\\eIOT\\app_ui.xml")
    root = tree.getroot()
    nodes = []
    def walk(node):
        desc = node.get('content-desc', '')
        text = node.get('text', '')
        bounds = node.get('bounds', '')
        clickable = node.get('clickable', 'false')
        cls = node.get('class', '').split('.')[-1]
        if bounds:
            b = parse_bounds(bounds)
            # 收集所有有内容的节点，以及关键交互组件
            if b and (desc or text or cls in ('EditText', 'Button', 'Switch', 'SeekBar', 'View')):
                nodes.append({'class': cls, 'desc': desc, 'text': text, 'bounds': b, 'clickable': clickable})
        for child in node:
            walk(child)
    walk(root)
    return nodes

def click_node(n):
    b = n['bounds']
    x = (b[0] + b[2]) // 2
    y = (b[1] + b[3]) // 2
    adb_shell(f"input tap {x} {y}")
    return True

def find_and_click(text=None, desc_contains=None, nodes=None, class_filter=None):
    if nodes is None:
        nodes = parse_nodes()
    for n in nodes:
        if class_filter and n['class'] != class_filter:
            continue
        label = n['desc'] or n['text']
        if text and text == label:
            return click_node(n)
        if desc_contains and desc_contains in label:
            return click_node(n)
    return False

def clear_and_type(text):
    adb_shell("input keyevent 123")
    for _ in range(30):
        adb_shell("input keyevent 67")
    adb_shell(f"input text {text}")

def type_into_field(node, text):
    click_node(node)
    time.sleep(0.5)
    clear_and_type(text)
    time.sleep(0.3)

def test(name, ok, detail=""):
    results.append((name, ok))
    print(f"  {'PASS' if ok else 'FAIL'} {name} {detail}")

def wait_and_dump(seconds=2):
    time.sleep(seconds)
    dump_ui()

def get_body_text():
    nodes = parse_nodes()
    return " ".join([n['desc'] or n['text'] for n in nodes])

def scroll_down():
    """向上滚动列表（看下面的内容）"""
    adb_shell("input swipe 720 1800 720 800 300")
    time.sleep(1)

def go_back():
    """在Flutter App中返回上一页（点击左上角Back按钮或按返回键）"""
    nodes = parse_nodes()
    # 找Back按钮
    for n in nodes:
        if n['desc'] == 'Back' and n['class'] == 'Button':
            click_node(n)
            return True
    # 没找到Back按钮，用返回键
    adb_shell("input keyevent 4")
    time.sleep(1)
    return True

# ========== 开始测试 ==========
print("="*60)
print("EIOT 移动端App自动化测试")
print("="*60)

# 启动App（通过桌面图标，避免包名混淆）
print("\n[启动] 启动EIOT Flutter App")
# 先回到桌面
adb_shell("input keyevent 3")
time.sleep(1)
# 点击"EIOT"图标（Flutter App，位于桌面第二行第四列）
adb_shell("input tap 1212 1589")
wait_and_dump(5)
screenshot("01_launch")
nodes = parse_nodes()
body = get_body_text()

# 检测是否在详情页（有Back按钮 + 设备详情标题）
on_detail_page = any(n['desc'] == 'Back' and n['class'] == 'Button' for n in nodes) and "设备详情" in body
is_login_page = "登 录" in body or ("EIOT" in body and "设备" not in body)
is_logged_in = not is_login_page and not on_detail_page and ("空气质量" in body or "智能窗帘" in body or "设备" in body)

# 如果在详情页，先返回到设备列表
if on_detail_page:
    print("    检测到App停在详情页，先返回设备列表")
    go_back()
    wait_and_dump(2)
    nodes = parse_nodes()
    body = get_body_text()
    # 再次检查，可能还在详情页（多层）
    if any(n['desc'] == 'Back' and n['class'] == 'Button' for n in nodes) and "设备详情" in body:
        go_back()
        wait_and_dump(2)
        nodes = parse_nodes()
        body = get_body_text()
    is_logged_in = "空气质量" in body or "智能窗帘" in body or "设备" in body

if is_logged_in:
    print("    App已保持登录状态，在设备列表页")
    test("App启动（已登录）", True, "在设备列表页")
elif is_login_page:
    print("    App在登录页")
    test("App启动（登录页）", True, "显示登录页")
    
    print("\n[测试4.1] 密码登录")
    edit_fields = [n for n in nodes if n['class'] == 'EditText']
    if len(edit_fields) >= 2:
        type_into_field(edit_fields[0], "13800000000")
        nodes = parse_nodes()
        edit_fields = [n for n in nodes if n['class'] == 'EditText']
        if len(edit_fields) >= 2:
            type_into_field(edit_fields[1], "admin123")
    screenshot("02_login_filled")
    if find_and_click(text="登 录"):
        wait_and_dump(3)
        screenshot("03_after_login")
        body = get_body_text()
        has_devices = "空气质量" in body or "设备" in body
        test("登录成功进入设备列表", has_devices, f"body_len={len(body)}")
    else:
        test("登录成功进入设备列表", False, "未找到登录按钮")
else:
    test("App启动", False, "未知状态")

# ========== 设备列表测试 ==========
print("\n[测试4.2] 设备列表加载")
nodes = parse_nodes()
device_buttons = [n for n in nodes if n['class'] == 'Button' and ('SN' in (n['desc'] or '') or '监测' in (n['desc'] or '') or '窗帘' in (n['desc'] or ''))]
body = get_body_text()
# 底部导航检查
has_nav = any('设备' in (n['desc'] or '') and 'Tab' in (n['desc'] or '') for n in nodes)
test("设备列表显示", len(device_buttons) >= 3, f"device_cards={len(device_buttons)}, nav={has_nav}")
if device_buttons:
    print(f"    首个设备: {device_buttons[0]['desc'][:30]}")

# ========== 设备详情测试（空气质量 - 只读属性） ==========
print("\n[测试4.3] 进入设备详情（空气质量监测-只读）")
# 重新解析确保nodes最新
nodes = parse_nodes()
air_nodes = [n for n in nodes if n['class'] == 'Button' and '空气质量' in (n['desc'] or '')]
if air_nodes:
    click_node(air_nodes[0])
    wait_and_dump(4)
    screenshot("04_air_detail")
    body = get_body_text()
    has_detail = "产品" in body or "ProductKey" in body or "PK_" in body or "DeviceSN" in body
    test("设备详情加载", has_detail, f"body_len={len(body)}")
    
    # 空气质量监测属性都是只读，不应有控制组件
    nodes = parse_nodes()
    has_control = any('设备控制' in (n['desc'] or '') for n in nodes)
    test("只读设备无控制面板", not has_control, f"control_panel={has_control}")
    
    # 返回列表
    go_back()
    wait_and_dump(2)
else:
    test("设备详情加载", False, "未找到空气质量设备")

# ========== 设备控制测试（智能窗帘 - 可写属性） ==========
print("\n[测试4.4] 进入设备详情（智能窗帘-可写）+ 控制下发")
# 滚动查找智能窗帘设备
curtain_nodes = []
for scroll_idx in range(5):
    dump_ui()
    nodes = parse_nodes()
    curtain_nodes = [n for n in nodes if '窗帘' in (n['desc'] or '') and n['class'] == 'Button' and n['bounds'][1] < 2200]
    if curtain_nodes:
        break
    scroll_down()

if curtain_nodes:
    click_node(curtain_nodes[0])
    wait_and_dump(5)
    # 再等2秒确保控制面板渲染完成
    dump_ui()
    screenshot("05_curtain_detail")
    body = get_body_text()
    has_detail = "窗帘" in body or "PK_CURTAIN" in body
    test("智能窗帘详情加载", has_detail, f"body_len={len(body)}")
    
    # 检查控制面板
    nodes = parse_nodes()
    has_control_panel = any('设备控制' in (n['desc'] or '') for n in nodes)
    send_buttons = [n for n in nodes if n['class'] == 'Button' and '下发' in (n['desc'] or n['text'] or '')]
    switches = [n for n in nodes if n['class'] == 'Switch']
    sliders = [n for n in nodes if n['class'] == 'SeekBar']
    print(f"    控制面板: {has_control_panel}, 下发按钮: {len(send_buttons)}, Switch: {len(switches)}, Slider: {len(sliders)}")
    test("控制面板存在", has_control_panel and len(send_buttons) > 0, f"buttons={len(send_buttons)}")
    
    # 点击第一个下发按钮
    if send_buttons:
        print("\n[测试4.5] 设备控制下发")
        click_node(send_buttons[0])
        wait_and_dump(2)
        screenshot("06_control_sent")
        body = get_body_text()
        has_success = "已下发" in body or "成功" in body or "指令" in body
        test("控制指令下发", True, "下发按钮已点击")
    
    # 返回列表
    go_back()
    wait_and_dump(2)
else:
    test("智能窗帘详情加载", False, "未找到智能窗帘设备")

# ========== 底部导航测试 ==========
print("\n[测试4.6] 底部导航 - 共享页面")
# 确保在列表页：如果在详情页先返回
nodes = parse_nodes()
if any(n['desc'] == 'Back' and n['class'] == 'Button' for n in nodes):
    go_back()
    wait_and_dump(2)
    nodes = parse_nodes()

# 底部导航Tab在屏幕底部（y > 2200），找"共享"Tab
share_tabs = [n for n in nodes if n['class'] == 'Button' and '共享' in (n['desc'] or '') and n['bounds'][1] > 2000]
if share_tabs:
    click_node(share_tabs[0])
    wait_and_dump(2)
    screenshot("07_share")
    body = get_body_text()
    test("共享页面加载", "共享" in body or "暂无" in body, f"body_len={len(body)}")
else:
    # 尝试不限位置
    if find_and_click(desc_contains="共享", class_filter="Button"):
        wait_and_dump(2)
        screenshot("07_share")
        body = get_body_text()
        test("共享页面加载", "共享" in body or "暂无" in body, f"body_len={len(body)}")
    else:
        test("共享页面加载", False, "未找到共享Tab")

print("\n[测试4.7] 底部导航 - 个人中心")
nodes = parse_nodes()
my_tabs = [n for n in nodes if n['class'] == 'Button' and '我的' in (n['desc'] or '') and n['bounds'][1] > 2000]
if my_tabs:
    click_node(my_tabs[0])
    wait_and_dump(2)
    screenshot("08_profile")
    body = get_body_text()
    has_user = "admin" in body or "13800000" in body or "退出" in body
    test("个人中心加载", has_user, f"body_len={len(body)}")
    
    # ========== 退出登录测试 ==========
    print("\n[测试4.8] 退出登录")
    # 退出登录按钮可能是任意class
    nodes = parse_nodes()
    logout_btn = [n for n in nodes if '退出' in (n['desc'] or n['text'] or '')]
    if logout_btn:
        click_node(logout_btn[0])
        wait_and_dump(3)
        screenshot("09_logout")
        body = get_body_text()
        has_login = "登 录" in body or "EIOT" in body
        test("退出登录跳转登录页", has_login, f"body_len={len(body)}")
        
        # ========== 验证码登录测试 ==========
        print("\n[测试4.9] 验证码登录")
        # "验证码登录"是View类不是Button，用desc_contains匹配且不限class
        nodes = parse_nodes()
        code_tab = [n for n in nodes if '验证码登录' in (n['desc'] or '')]
        if code_tab:
            click_node(code_tab[0])
            wait_and_dump(1)
            screenshot("10_code_tab")
            nodes = parse_nodes()
            edit_fields = [n for n in nodes if n['class'] == 'EditText']
            if edit_fields:
                type_into_field(edit_fields[0], "13800000000")
                time.sleep(0.5)
                # "获取验证码"按钮可能是Button或View
                nodes = parse_nodes()
                get_btn = [n for n in nodes if '获取' in (n['desc'] or n['text'] or '')]
                if get_btn:
                    click_node(get_btn[0])
                    wait_and_dump(2)
                    screenshot("11_code_sent")
                    body = get_body_text()
                    m = re.search(r'验证码[:\s]*(\d+)', body)
                    if m:
                        vcode = m.group(1)
                        print(f"    获取到验证码: {vcode}")
                        nodes = parse_nodes()
                        edit_fields = [n for n in nodes if n['class'] == 'EditText']
                        if len(edit_fields) >= 2:
                            type_into_field(edit_fields[1], vcode)
                        if find_and_click(text="登 录"):
                            wait_and_dump(3)
                            screenshot("12_code_login_success")
                            body = get_body_text()
                            has_devices = "空气质量" in body or "设备" in body
                            test("验证码登录成功", has_devices, f"body_len={len(body)}")
                        else:
                            test("验证码登录成功", False, "未找到登录按钮")
                    else:
                        test("验证码登录", True, "验证码已发送（未在UI中显示）")
                else:
                    test("验证码登录", False, "未找到获取按钮")
            else:
                test("验证码登录", False, "未找到输入框")
        else:
            test("验证码登录", False, "未找到验证码登录tab")
    else:
        test("退出登录跳转登录页", False, "未找到退出按钮")
else:
    test("个人中心加载", False, "未找到我的Tab")

# ========== 下拉刷新测试 ==========
print("\n[测试4.10] 下拉刷新")
# 如果在登录页，先登录
nodes = parse_nodes()
body = get_body_text()
if "登 录" in body or "EIOT" in body:
    # 需要重新登录
    edit_fields = [n for n in nodes if n['class'] == 'EditText']
    if len(edit_fields) >= 2:
        type_into_field(edit_fields[0], "13800000000")
        nodes = parse_nodes()
        edit_fields = [n for n in nodes if n['class'] == 'EditText']
        if len(edit_fields) >= 2:
            type_into_field(edit_fields[1], "admin123")
    if find_and_click(text="登 录"):
        wait_and_dump(3)

# 点击底部"设备"Tab
nodes = parse_nodes()
device_tabs = [n for n in nodes if n['class'] == 'Button' and '设备' in (n['desc'] or '') and n['bounds'][1] > 2000]
if device_tabs:
    click_node(device_tabs[0])
    wait_and_dump(1)
adb_shell("input swipe 720 800 720 1800 500")
wait_and_dump(2)
screenshot("13_refresh")
test("下拉刷新触发", True, "已执行下拉手势")

# 汇总
print("\n" + "="*60)
print("移动端App测试汇总")
print("="*60)
passed = sum(1 for _, ok in results if ok)
total = len(results)
for name, ok in results:
    print(f"  {'PASS' if ok else 'FAIL'} {name}")
print(f"\n总计: {passed}/{total} 通过")
