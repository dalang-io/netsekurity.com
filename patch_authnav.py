#!/usr/bin/env python3
p = "static/index.html"
s = open(p).read()

# Header nav buttons -> placeholder (server injects based on auth)
old_hdr = '''      <div class="flex items-center gap-3">
        <a href="/dashboard" class="prompt hidden text-[13px] text-gray-400 hover:text-emerald-300 sm:inline">login</a>
        <a href="/dashboard" class="rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-[13px] font-bold text-emerald-300 hover:bg-emerald-500/20 glow">
          ./dashboard<span class="cursor"></span>
        </a>
      </div>'''
new_hdr = '''      <div class="flex items-center gap-3">
        __HEADER_AUTH_NAV__
      </div>'''
assert old_hdr in s, "header nav anchor"
s = s.replace(old_hdr, new_hdr, 1)

# Footer login -> placeholder
old_ft = '''        <a href="/dashboard" class="hover:text-emerald-300">login</a>'''
new_ft = '''        __FOOTER_AUTH_NAV__'''
assert old_ft in s, "footer login anchor"
s = s.replace(old_ft, new_ft, 1)

open(p, "w").write(s)
print("index.html auth-nav placeholders added")
