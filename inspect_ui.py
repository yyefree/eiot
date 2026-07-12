import xml.etree.ElementTree as ET
import sys
tree = ET.parse(sys.argv[1] if len(sys.argv) > 1 else 'd:\\AI\\eIOT\\app_ui_check.xml')
root = tree.getroot()
for n in root.iter('node'):
    cls = n.get('class', '').split('.')[-1]
    text = n.get('text', '')
    desc = n.get('content-desc', '')
    bounds = n.get('bounds', '')
    if text or desc or cls in ('EditText', 'Button', 'Switch', 'SeekBar'):
        print(f"cls={cls} text={text!r} desc={desc!r} bounds={bounds}")
