import codecs
with codecs.open('svc/tmaxupgrade.go', 'r', 'utf-16') as f:
    content = f.read()
with codecs.open('svc/tmaxupgrade.go', 'w', 'utf-8') as f:
    f.write(content)
