#!/usr/bin/env python3
p = "dashboard.go"
s = open(p).read()

fixes = [
    # grid wrapper
    ('class="grid gap-4 lg:grid-cols-2"',
     'class="grid w-full gap-4 lg:grid-cols-2'),
    # left column space-y-4
    ('<div class="space-y-4">',
     '<div class="min-w-0 w-full space-y-4">'),
    # right column (domains section parent) — the h-fit section is a direct grid child
    ('<section class="rounded border border-cyan-500/30 bg-[#04060c] h-fit">',
     '<section class="min-w-0 w-full rounded border border-cyan-500/30 bg-[#04060c] h-fit">'),
    # balance section grid child
    ('<section class="rounded border border-emerald-500/30 bg-[#04060c]">',
     '<section class="min-w-0 w-full rounded border border-emerald-500/30 bg-[#04060c]">'),
    # history section grid child
    ('<section class="rounded border border-white/10 bg-[#04060c]">',
     '<section class="min-w-0 w-full rounded border border-white/10 bg-[#04060c]">'),
    # topup section grid child (same white/10 border — but that's also history; the NEXT one)
    # topup is 'rounded border border-white/10 bg-[#04060c]' too; both get min-w-0 (fine)
    # main: ensure no overflow
    ('<main id="main" class="mx-auto max-w-6xl px-4 py-5 sm:px-6">',
     '<main id="main" class="mx-auto max-w-6xl overflow-x-hidden px-4 py-5 sm:px-6">'),
    # domain add form row
    ('<form class="flex gap-2" hx-post="/api/domains"',
     '<form class="flex min-w-0 gap-2" hx-post="/api/domains"'),
]

for old, new in fixes:
    if old in s:
        s = s.replace(old, new, 1)
    else:
        print("WARN not found:", old[:60])

open(p, "w").write(s)
print("dashboard.go min-w-0 overflow hardening done")