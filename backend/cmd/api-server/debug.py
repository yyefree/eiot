import re

with open('main.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the log.Println line
idx = content.find('log.Println("[EMQX] subscribed Feiyan-style topics:')
if idx >= 0:
    # Show context
    print('Context around index:')
    print(repr(content[idx-50:idx+200]))
else:
    print("Not found")