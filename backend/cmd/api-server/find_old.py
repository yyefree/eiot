import re

with open('main.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the log.Println line
idx = content.find('log.Println("[EMQX] subscribed Feiyan-style topics:')
if idx >= 0:
    # Go backwards to find the closing brace
    start = content.rfind('\t\t})', 0, idx)
    if start >= 0:
        old = content[start:idx+200]
        print('OLD (first 500 chars):')
        print(repr(old[:500]))
    else:
        print('Could not find start brace')
else:
    print('Not found')