import xml.etree.ElementTree as ET
import sys

tree = ET.parse(sys.argv[1] if len(sys.argv) > 1 else 'd:\\AI\\eIOT\\test_ui.xml')
root = tree.getroot()

def walk(node, depth=0):
    desc = node.get('content-desc', '')
    text = node.get('text', '')
    bounds = node.get('bounds', '')
    clickable = node.get('clickable', 'false')
    cls = node.get('class', '').split('.')[-1]
    if cls in ('EditText', 'Button', 'View', 'Switch', 'SeekBar') and bounds:
        label = (desc or text or '').replace('&#10;', ' | ')[:80]
        print(f"{'  '*depth}[{cls}] bounds={bounds} click={clickable} desc={label}")
    for child in node:
        walk(child, depth+1)

walk(root)
