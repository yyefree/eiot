import xml.etree.ElementTree as ET
tree = ET.parse('d:\\AI\\eIOT\\app_ui_login.xml')
root = tree.getroot()
for n in root.iter('node'):
    cls = n.get('class', '').split('.')[-1]
    text = n.get('text', '')
    desc = n.get('content-desc', '')
    bounds = n.get('bounds', '')
    if text or desc or cls in ('EditText', 'Button', 'Switch', 'SeekBar', 'TextView'):
        print(f"cls={cls} text={text!r} desc={desc!r} bounds={bounds}")
