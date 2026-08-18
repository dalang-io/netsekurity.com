#!/usr/bin/env python3
p = "main.go"
s = open(p).read()
old = '''\ts := strings.ReplaceAll(string(b), "__GOOGLE_CLIENT_ID__", getenv("GOOGLE_CLIENT_ID", ""))
\ts = strings.ReplaceAll(s, "__CSS_HASH__", cssHash)
\tw.Header().Set("Content-Type", "text/html; charset=utf-8")
\tw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
\tw.Write([]byte(s))'''
new = '''\ts := strings.ReplaceAll(string(b), "__GOOGLE_CLIENT_ID__", getenv("GOOGLE_CLIENT_ID", ""))
\ts = strings.ReplaceAll(s, "__CSS_HASH__", cssHash)
\t// Auth-aware nav: logged-in shows dashboard, anonymous shows login.
\t_, auErr := currentUser(r)
\thdrNav := `<a href="/dashboard" class="rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-[13px] font-bold text-emerald-300 hover:bg-emerald-500/20 glow">./dashboard<span class="cursor"></span></a>`
\tloginNav := `<a href="/login" class="rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-[13px] font-bold text-emerald-300 hover:bg-emerald-500/20 glow">login</a>`
\tif auErr == nil {
\t\ts = strings.ReplaceAll(s, "__HEADER_AUTH_NAV__", hdrNav)
\t\ts = strings.ReplaceAll(s, "__FOOTER_AUTH_NAV__", `<a href="/dashboard" class="hover:text-emerald-300">dashboard</a>`)
\t} else {
\t\ts = strings.ReplaceAll(s, "__HEADER_AUTH_NAV__", loginNav)
\t\ts = strings.ReplaceAll(s, "__FOOTER_AUTH_NAV__", `<a href="/login" class="hover:text-emerald-300">login</a>`)
\t}
\tw.Header().Set("Content-Type", "text/html; charset=utf-8")
\tw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
\tw.Write([]byte(s))'''
assert old in s, "handleIndex tail anchor"
s = s.replace(old, new, 1)
open(p, "w").write(s)
print("main.go handleIndex auth-nav injected")