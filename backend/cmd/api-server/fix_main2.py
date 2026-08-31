import re

with open('main.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the log.Println line
idx = content.find('log.Println("[EMQX] subscribed Feiyan-style topics:')
if idx >= 0:
    print('Found at index:', idx)
    # Show the surrounding context
    print(content[idx:idx+200])
else:
    print('Not found')