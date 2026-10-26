#!/usr/bin/env python3
"""
Generates the isometric illustrations for netsekurity.com and writes site_svg.go.

    python3 tools/svg/iso.py            # rewrites site_svg.go
    python3 tools/svg/iso.py --preview  # also writes a standalone preview page

Why code and not a drawing tool: every solid is drawn in one projection (2:1
dimetric, x = (X - Y) * u, y = ((X + Y) / 2 - Z) * u) with three flat tones (top
lightest, left mid, right darkest) and a thick ink outline, the same system
virus.ai.id uses, so the two sites read as one family. Changing a colour or a
proportion is an edit here, not a redraw.

Rules the drawings obey (see CLAIMS.md and README "Public site"):
  - Decorative. Each is aria-hidden; the meaning is always also in HTML text, so
    nothing is carried by a drawing alone and nothing needs translating.
  - No words inside, only the digits 1-5 on the flow diagram, which are language
    neutral and large enough (>= 11px at phone width, see FLOW_W).
  - No motion. Nothing loops, nothing animates (WCAG 2.2.2).
  - Sober subjects: layers, servers, documents, locks. No skulls, hackers or maps.
"""
import sys

INK = "#1c1a17"
SW = 1.5  # outline width

# three tones per material: (top, left, right)
YELLOW = ("#fde047", "#facc15", "#eab308")
WHITE = ("#ffffff", "#f3f4f6", "#e5e7eb")
GRAY = ("#e5e7eb", "#d1d5db", "#9ca3af")
DARK = ("#4b5563", "#374151", "#1f2937")
GREEN = ("#bbf7d0", "#86efac", "#4ade80")
RED = ("#fecaca", "#fca5a5", "#f87171")
AMBER = ("#fde68a", "#fcd34d", "#f59e0b")
BLUE = ("#bae6fd", "#7dd3fc", "#38bdf8")
CREAM = ("#fffbeb", "#fef3c7", "#fde68a")


class Scene:
    def __init__(self, u, prefix):
        self.u, self.prefix = u, prefix
        self.items = []  # (layer, depth, svg)
        self.layer = 0   # raise it for anything standing on / above what came before
        self.pts = []
        self.defs = []

    # ---- projection -------------------------------------------------------
    def P(self, X, Y, Z=0.0):
        x, y = (X - Y) * self.u, ((X + Y) / 2.0 - Z) * self.u
        self.pts.append((x, y))
        return x, y

    def _fmt(self, pts):
        return " ".join("%.1f,%.1f" % p for p in pts)

    def add(self, depth, svg):
        self.items.append((self.layer, depth, svg))

    # ---- primitives ------------------------------------------------------
    def poly(self, pts3, fill, depth, stroke=INK, sw=SW, extra=""):
        pts = [self.P(*p) for p in pts3]
        self.add(depth, '<polygon points="%s" fill="%s" stroke="%s" stroke-width="%s" stroke-linejoin="round"%s/>'
                 % (self._fmt(pts), fill, stroke, sw, (" " + extra) if extra else ""))

    def box(self, X, Y, Z, w, d, h, mat, top=None, left=None, right=None, sw=SW):
        """Draw a solid. mat is a (top, left, right) tuple; the keyword overrides win."""
        t, l, r = mat
        t, l, r = top or t, left or l, right or r
        depth = X + Y + w + d + Z * 0.01
        # left face: plane Y = Y+d ; right face: plane X = X+w ; top: Z = Z+h
        self.poly([(X, Y + d, Z), (X + w, Y + d, Z), (X + w, Y + d, Z + h), (X, Y + d, Z + h)], l, depth, sw=sw)
        self.poly([(X + w, Y, Z), (X + w, Y + d, Z), (X + w, Y + d, Z + h), (X + w, Y, Z + h)], r, depth, sw=sw)
        self.poly([(X, Y, Z + h), (X + w, Y, Z + h), (X + w, Y + d, Z + h), (X, Y + d, Z + h)], t, depth, sw=sw)
        return depth

    def top_rect(self, X, Y, Z, w, d, fill, depth, stroke="none", sw=0, opacity=None):
        """A flat rectangle lying on a horizontal plane at height Z (marks on a top face)."""
        pts = [self.P(X, Y, Z), self.P(X + w, Y, Z), self.P(X + w, Y + d, Z), self.P(X, Y + d, Z)]
        extra = ' fill-opacity="%s"' % opacity if opacity is not None else ""
        self.add(depth, '<polygon points="%s" fill="%s" stroke="%s" stroke-width="%s"%s/>' % (self._fmt(pts), fill, stroke, sw, extra))

    def line3(self, a, b, depth, stroke=INK, sw=SW, dash=None, cap="round"):
        p, q = self.P(*a), self.P(*b)
        d = ' stroke-dasharray="%s"' % dash if dash else ""
        self.add(depth, '<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%s" stroke-linecap="%s"%s/>'
                 % (p[0], p[1], q[0], q[1], stroke, sw, cap, d))

    def polyline3(self, pts3, depth, stroke=INK, sw=SW, fill="none", close=False):
        pts = [self.P(*p) for p in pts3]
        tag = "polygon" if close else "polyline"
        self.add(depth, '<%s points="%s" fill="%s" stroke="%s" stroke-width="%s" stroke-linecap="round" stroke-linejoin="round"/>'
                 % (tag, self._fmt(pts), fill, stroke, sw))

    def disc_top(self, X, Y, Z, r, fill, depth, stroke=INK, sw=SW):
        """A circle lying on a horizontal plane: an exact iso ellipse via the plane's matrix."""
        u = self.u
        self.P(X - r, Y - r, Z), self.P(X + r, Y + r, Z), self.P(X - r, Y + r, Z), self.P(X + r, Y - r, Z)
        self.add(depth, '<g transform="matrix(%.3f %.3f %.3f %.3f 0 %.3f)"><circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" stroke="%s" stroke-width="%s" vector-effect="non-scaling-stroke"/></g>'
                 % (u, u / 2.0, -u, u / 2.0, -Z * u, X, Y, r, fill, stroke, sw))

    def sphere(self, X, Y, Z, r, mat, depth):
        """A ball: a circle with a lighter highlight, placed at the projected centre."""
        cx, cy = self.P(X, Y, Z)
        R = r * self.u
        self.pts += [(cx - R, cy - R), (cx + R, cy + R)]
        self.add(depth, '<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" stroke="%s" stroke-width="%s"/>'
                 '<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" opacity=".55"/>'
                 % (cx, cy, R, mat[1], INK, SW, cx - R * .3, cy - R * .32, R * .32, mat[0]))

    def digit(self, n, X, Y, Z, depth):
        cx, cy = self.P(X, Y, Z)
        self.pts += [(cx - 13, cy - 13), (cx + 13, cy + 13)]
        self.add(depth, '<circle cx="%.1f" cy="%.1f" r="12" fill="%s" stroke="%s" stroke-width="%s"/>'
                 '<text x="%.1f" y="%.1f" text-anchor="middle" font-family="Plus Jakarta Sans, system-ui, sans-serif" font-weight="800" font-size="15" fill="%s">%s</text>'
                 % (cx, cy, YELLOW[1], INK, SW, cx, cy + 5.3, INK, n))

    def arrow_down(self, X, Y, Z0, Z1, depth, stroke=INK):
        """A vertical arrow in the scene (screen-vertical), head at the lower end."""
        self.line3((X, Y, Z0), (X, Y, Z1 + 0.35), depth, stroke=stroke, sw=2.2)
        x, y = self.P(X, Y, Z1)
        self.add(depth, '<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="%s" stroke="%s" stroke-width="1.5" stroke-linejoin="round"/>'
                 % (x - 7, y - 9, x + 7, y - 9, x, y + 1, stroke, stroke))

    def arrow_screen(self, x, y0, y1, depth, stroke=INK):
        """A straight arrow on screen from y0 down to y1 at column x (a chain of nodes)."""
        self.pts += [(x - 8, y0), (x + 8, y1 + 2)]
        self.add(depth, '<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="2.4" stroke-linecap="round"/>'
                 '<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="%s" stroke="%s" stroke-width="1.5" stroke-linejoin="round"/>'
                 % (x, y0, x, y1 - 5, stroke, x - 7, y1 - 10, x + 7, y1 - 10, x, y1 + 1, stroke, stroke))

    # ---- output ----------------------------------------------------------
    def svg(self, pad=8, maxw=None):
        xs = [p[0] for p in self.pts]
        ys = [p[1] for p in self.pts]
        x0, y0, x1, y1 = min(xs) - pad, min(ys) - pad, max(xs) + pad, max(ys) + pad
        w, h = x1 - x0, y1 - y0
        body = "".join(it[2] for it in sorted(self.items, key=lambda it: (it[0], it[1])))
        defs = "<defs>%s</defs>" % "".join(self.defs) if self.defs else ""
        return ('<svg xmlns="http://www.w3.org/2000/svg" viewBox="%.0f %.0f %.0f %.0f" width="%.0f" height="%.0f" '
                'class="h-auto w-full" aria-hidden="true" focusable="false">%s%s</svg>'
                % (x0, y0, w, h, w, h, defs, body)), w, h


def scope_scene():
    """What is scanned and what is not: three layers pulled apart. Yellow is
    scanned from outside; the two grey layers (behind a login; data and business
    logic) are not."""
    s = Scene(20, "scope")
    s.defs.append('<pattern id="scope-hatch" width="7" height="7" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">'
                  '<rect width="7" height="7" fill="#e5e7eb"/><line x1="0" y1="0" x2="0" y2="7" stroke="#9ca3af" stroke-width="2"/></pattern>')
    W = 7.0
    s.layer = -1
    # dashed guides joining the layers (exploded view)
    for (gx, gy) in [(0, 0), (W, 0), (W, W), (0, W)]:
        s.line3((gx, gy, 0.5), (gx, gy, 4.8), 0, stroke="#6b7280", sw=1.2, dash="3 4")
    # bottom layer: data and business logic (not scanned)
    s.layer = 0
    s.box(0, 0, 0, W, W, 0.5, GRAY, top="url(#scope-hatch)")
    s.layer = 1
    for i in range(3):  # a stack of slabs reads as a database
        s.box(2.4, 2.4, 0.5 + i * 0.5, 2.2, 2.2, 0.38, DARK if i == 1 else WHITE)
    # middle layer: behind a login (not scanned)
    s.layer = 2
    s.box(0, 0, 2.4, W, W, 0.5, GRAY, top="url(#scope-hatch)")
    s.layer = 3
    z = 2.9
    s.box(2.7, 2.8, z, 1.8, 1.3, 1.1, DARK)             # lock body
    s.box(2.95, 3.0, z + 1.1, 0.32, 0.3, 0.8, GRAY)     # shackle, left post
    s.box(3.95, 3.0, z + 1.1, 0.32, 0.3, 0.8, GRAY)     # shackle, right post
    s.box(2.95, 3.0, z + 1.9, 1.32, 0.3, 0.32, GRAY)    # shackle, top bar
    # top layer: the public surface (scanned)
    top_z = 4.8
    s.layer = 4
    s.box(0, 0, top_z, W, W, 0.5, YELLOW)
    s.layer = 5
    z = top_z + 0.5
    s.box(0.8, 0.9, z, 3.2, 2.2, 1.3, WHITE, top=YELLOW[0])                    # the website
    for i in range(3):                                                          # exposed files
        s.box(4.9, 1.0 + i * 0.85, z, 1.3, 0.6, 0.18 + i * 0.12, WHITE)
    s.box(1.0, 4.5, z, 0.55, 2.0, 1.7, GRAY)                                    # header wall
    s.box(2.4, 4.6, z, 1.5, 1.5, 0.35, GREEN)                                   # certificate
    s.box(4.7, 4.5, z, 1.2, 1.2, 0.9, AMBER)                                    # redirect marker
    # the scanner, outside, reading from the left, and its read-only beam
    s.layer = 6
    sx, sy = -3.4, 2.6
    s.layer = 5
    s.box(sx - 0.5, sy - 0.5, z - 0.3, 2.1, 2.1, 0.3, GRAY)   # a pad, so it stands on something
    s.layer = 6
    s.box(sx, sy, z, 1.1, 1.1, 1.9, DARK)
    s.top_rect(sx + 0.2, sy + 0.2, z + 1.9, 0.7, 0.7, "#facc15", 900, INK, 1.2)
    beam = [(sx + 1.1, sy + 0.2, z + 1.2), (-0.3, 0.4, z + 0.9), (-0.3, 3.4, z + 0.9), (sx + 1.1, sy + 0.9, z + 1.2)]
    s.poly(beam, "#facc15", 800, stroke="#a16207", sw=1.2, extra='fill-opacity=".38" stroke-dasharray="5 4"')
    return s.svg(pad=10)


def report_scene():
    """A report: three sheets, the top one showing a heading, severity chips and
    rows of findings."""
    s = Scene(18, "report")
    a, b = 6.0, 7.6  # sheet width (X) and depth (Y)
    for i, (dx, dy, z) in enumerate([(0.0, 0.0, 0.0), (0.0, 0.0, 0.5), (0.0, 0.0, 1.0)]):
        s.box(dx, dy, z, a, b, 0.3, WHITE if i else GRAY)
    zt = 1.3
    d = 500
    s.top_rect(0.5, 0.5, zt, 5.0, 0.9, "#facc15", d, INK, 1.2)                  # heading bar
    chips = [("#fecaca", 0.5), ("#fde68a", 1.5), ("#bae6fd", 2.5), ("#e5e7eb", 3.5)]
    for fill, x in chips:                                                       # severity chips
        s.top_rect(x, 1.8, zt, 0.8, 0.55, fill, d, INK, 1.1)
    for i in range(5):                                                          # finding rows
        y = 2.8 + i * 0.9
        s.top_rect(0.5, y, zt, 0.55, 0.55, ["#fca5a5", "#fcd34d", "#fcd34d", "#7dd3fc", "#d1d5db"][i], d, INK, 1.0)
        s.top_rect(1.3, y + 0.12, zt, 4.1 - i * 0.35, 0.28, "#d1d5db", d)
    return s.svg(pad=10)


def step_ownership():
    s = Scene(15, "s1")
    s.box(0, 0, 0, 3.2, 3.2, 1.5, WHITE, top=YELLOW[0])
    s.top_rect(0.5, 0.5, 1.5, 2.2, 0.55, "#ffffff", 400, INK, 1.1)
    s.top_rect(0.5, 1.4, 1.5, 1.4, 0.55, "#d1d5db", 400)
    s.box(2.6, 2.6, 1.5, 0.22, 0.22, 2.5, DARK)                                 # pole
    s.poly([(2.82, 2.7, 3.95), (2.82, 2.7, 3.05), (2.82, 1.4, 3.5)], "#facc15", 900)  # flag
    return s.svg(pad=8)


def step_start():
    s = Scene(15, "s2")
    s.box(0, 0, 0, 4.4, 3.0, 0.7, GRAY)
    s.layer = 1
    s.box(0.4, 0.4, 0.7, 3.6, 2.2, 0.9, YELLOW)
    s.polyline3([(1.5, 0.9, 1.6), (1.5, 2.1, 1.6), (2.7, 1.5, 1.6)], 400, stroke=INK, sw=2, fill=INK, close=True)
    return s.svg(pad=8)


def step_scan():
    s = Scene(15, "s3")
    s.box(0, 0, 0, 1.8, 1.8, 2.0, DARK)
    s.top_rect(0.3, 0.3, 2.0, 1.2, 1.2, "#facc15", 400, INK, 1.1)
    s.poly([(1.8, 0.3, 1.4), (3.3, 0.5, 1.1), (3.3, 2.6, 1.1), (1.8, 1.5, 1.4)], "#facc15", 300,
           stroke="#a16207", sw=1.1, extra='fill-opacity=".4" stroke-dasharray="4 3"')
    s.box(3.4, 0.6, 0, 1.9, 1.9, 1.2, YELLOW)
    s.polyline3([(3.9, 1.5, 1.2), (4.3, 1.9, 1.2), (5.0, 1.0, 1.2)], 900, stroke=INK, sw=2.4)
    return s.svg(pad=8)


def step_report():
    s = Scene(15, "s4")
    s.box(0, 0, 0, 4.2, 5.2, 0.4, GRAY)
    s.box(0, 0, 0.4, 4.2, 5.2, 0.3, WHITE)
    s.top_rect(0.4, 0.4, 0.7, 3.4, 0.6, "#facc15", 400, INK, 1.0)
    for i in range(3):
        s.top_rect(0.4, 1.4 + i * 0.75, 0.7, 3.0 - i * 0.5, 0.35, "#d1d5db", 400)
    s.disc_top(3.2, 4.2, 0.7, 0.65, "#facc15", 500)
    s.polyline3([(2.95, 4.15, 0.72), (3.2, 4.45, 0.72), (3.5, 3.9, 0.72)], 600, stroke=INK, sw=2)
    return s.svg(pad=8)


def tiers_scene():
    """Three steps up: automated, expert review, whitebox."""
    s = Scene(18, "tiers")
    s.box(0, 0, 0, 2.6, 3.6, 1.0, YELLOW)
    s.box(2.6, 0, 0, 2.6, 3.6, 2.0, WHITE, top=YELLOW[0])
    s.box(5.2, 0, 0, 2.6, 3.6, 3.0, DARK)
    s.layer = 1
    # a small marker on each step
    s.top_rect(0.5, 0.9, 1.0, 1.6, 0.5, "#ffffff", 300, INK, 1.1)
    s.top_rect(3.1, 0.9, 2.0, 1.6, 0.5, "#facc15", 300, INK, 1.1)
    # a person on the top step (body, head)
    s.layer = 2
    s.box(6.0, 1.3, 3.0, 1.0, 1.0, 1.2, YELLOW)
    s.sphere(6.5, 1.8, 4.65, 0.55, YELLOW, 2000)
    return s.svg(pad=10)


def flow_scene():
    """Trust page: how data moves. Top to bottom: your website, the scanner, the AI
    analysis agent, the report, you. The digits match the numbered legend in HTML."""
    s = Scene(15, "flow")
    pitch = 6.2                    # nodes sit on the screen-vertical diagonal X = Y
    heights = [1.3, 1.5, 1.7, 0.9, 2.5]
    t = [0, pitch, 2 * pitch, 3 * pitch, 4 * pitch + 1.6]   # the last gap is longer: you are taller than the report
    # 1 your website
    X = Y = t[0]
    s.box(X, Y, 0, 2.4, 2.4, 1.3, WHITE, top=YELLOW[0])
    s.top_rect(X + 0.4, Y + 0.4, 1.3, 1.6, 0.4, "#ffffff", 5, INK, 1.0)
    # 2 the scanner
    X = Y = t[1]
    s.box(X, Y, 0, 2.4, 2.4, 1.5, DARK)
    s.top_rect(X + 0.4, Y + 0.4, 1.5, 1.6, 0.35, "#facc15", 50, INK, 1.0)
    s.top_rect(X + 0.4, Y + 1.3, 1.5, 1.6, 0.35, "#facc15", 50, INK, 1.0)
    # 3 the AI analysis agent
    X = Y = t[2]
    s.box(X, Y, 0, 2.4, 2.4, 1.1, YELLOW)
    s.sphere(X + 1.2, Y + 1.2, 1.1 + 0.55, 0.6, GREEN, 5000)
    # 4 the report, locked
    X = Y = t[3]
    s.box(X, Y, 0, 2.6, 2.6, 0.35, WHITE)
    s.top_rect(X + 0.35, Y + 0.3, 0.35, 1.9, 0.4, "#facc15", 400, INK, 1.0)
    s.top_rect(X + 0.35, Y + 0.95, 0.35, 1.9, 0.25, "#d1d5db", 400)
    s.top_rect(X + 0.35, Y + 1.45, 0.35, 1.3, 0.25, "#d1d5db", 400)
    s.box(X + 1.6, Y + 1.5, 0.35, 0.9, 0.7, 0.55, DARK)
    # 5 you
    X = Y = t[4]
    s.box(X + 0.55, Y + 0.55, 0, 1.3, 1.3, 1.3, YELLOW)
    s.sphere(X + 1.2, Y + 1.2, 1.3 + 0.55, 0.6, YELLOW, 9000)
    # arrows down the chain: from the foot of one node to the top of the next
    for i in range(4):
        y0 = s.P(t[i] + 2.6, t[i] + 2.6, 0)[1] + 2
        y1 = s.P(t[i + 1], t[i + 1], heights[i + 1])[1] - 3
        s.arrow_screen(0, y0, y1, 200 + i)
    # numbers sit to the left of each node
    for i in range(5):
        s.digit(i + 1, t[i], t[i] + 3.6, 0, 9500 + i)
    return s.svg(pad=10)


SCENES = [
    ("svg-scope", scope_scene), ("svg-report", report_scene), ("svg-tiers", tiers_scene), ("svg-flow", flow_scene),
    ("svg-step-1", step_ownership), ("svg-step-2", step_start), ("svg-step-3", step_scan), ("svg-step-4", step_report),
]


def build():
    out, sizes = [], {}
    for name, fn in SCENES:
        svg, w, h = fn()
        sizes[name] = (round(w), round(h))
        out.append('{{define "%s"}}%s{{end}}' % (name, svg))
    return "\n".join(out), sizes


def main():
    body, sizes = build()
    go = ("package main\n\n"
          "// Code generated by tools/svg/iso.py; DO NOT EDIT. Edit the generator and re-run it.\n"
          "// Isometric illustrations: decorative (aria-hidden), no baked-in words, no motion.\n\n"
          "const siteSVGHTML = `\n" + body + "\n`\n")
    if "--check" in sys.argv:  # used by TestSVGFileIsGeneratedFromTheTool: no write, exit 1 on drift
        sys.exit(0 if open("site_svg.go").read() == go else 1)
    open("site_svg.go", "w").write(go)
    print("site_svg.go:", sizes)
    if "--preview" in sys.argv:
        page = ("<!doctype html><meta charset=utf-8><body style='font-family:system-ui;background:#fffbeb;padding:24px;display:grid;"
                "grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:24px'>" +
                "".join("<figure style='margin:0;background:#fff;border:2px solid #000;padding:12px'><div style='max-width:%dpx;margin:auto'>%s</div>"
                        "<figcaption style='font:600 13px system-ui;margin-top:8px'>%s</figcaption></figure>"
                        % (min(w, 420), fn()[0], n) for (n, fn), (w, h) in zip(SCENES, sizes.values())) + "</body>")
        open(sys.argv[sys.argv.index("--preview") + 1] if len(sys.argv) > sys.argv.index("--preview") + 1 else "svg-preview.html", "w").write(page)


if __name__ == "__main__":
    main()
