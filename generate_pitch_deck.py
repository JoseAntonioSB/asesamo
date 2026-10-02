from pptx import Presentation
from pptx.util import Inches, Pt, Emu
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN
from pptx.util import Inches, Pt
import copy

# ── Color Palette ──────────────────────────────────────────────────
BG_DARK      = RGBColor(0x0D, 0x1B, 0x2A)   # navy dark
BG_SLIDE     = RGBColor(0x11, 0x26, 0x3D)   # slide bg
ACCENT_GOLD  = RGBColor(0xF5, 0xA6, 0x23)   # golden amber
ACCENT_GREEN = RGBColor(0x2E, 0xCC, 0x71)   # emerald
ACCENT_RED   = RGBColor(0xE7, 0x4C, 0x3C)   # red
WHITE        = RGBColor(0xFF, 0xFF, 0xFF)
LIGHT_GRAY   = RGBColor(0xCC, 0xD6, 0xE0)
MID_BLUE     = RGBColor(0x1A, 0x87, 0xC8)

SLIDE_W = Inches(13.33)
SLIDE_H = Inches(7.5)

prs = Presentation()
prs.slide_width  = SLIDE_W
prs.slide_height = SLIDE_H

BLANK = prs.slide_layouts[6]   # completely blank


# ── Helper utilities ────────────────────────────────────────────────

def add_rect(slide, l, t, w, h, fill_color=None, line_color=None, line_width=None):
    shape = slide.shapes.add_shape(1, Inches(l), Inches(t), Inches(w), Inches(h))
    shape.line.fill.background()
    if fill_color:
        shape.fill.solid()
        shape.fill.fore_color.rgb = fill_color
    else:
        shape.fill.background()
    if line_color:
        shape.line.color.rgb = line_color
        if line_width:
            shape.line.width = line_width
    else:
        shape.line.fill.background()
    return shape


def add_textbox(slide, text, l, t, w, h,
                font_size=18, bold=False, color=WHITE,
                align=PP_ALIGN.LEFT, italic=False, wrap=True):
    txb = slide.shapes.add_textbox(Inches(l), Inches(t), Inches(w), Inches(h))
    txb.word_wrap = wrap
    tf  = txb.text_frame
    tf.word_wrap = wrap
    p   = tf.paragraphs[0]
    p.alignment = align
    run = p.add_run()
    run.text = text
    run.font.size  = Pt(font_size)
    run.font.bold  = bold
    run.font.color.rgb = color
    run.font.italic = italic
    return txb


def bg(slide, color=BG_SLIDE):
    add_rect(slide, 0, 0, 13.33, 7.5, fill_color=color)


def header_bar(slide, color=ACCENT_GOLD, height=0.08):
    add_rect(slide, 0, 0, 13.33, height, fill_color=color)


def slide_title(slide, title, subtitle=None):
    add_textbox(slide, title, 0.4, 0.15, 12.5, 0.7,
                font_size=28, bold=True, color=WHITE, align=PP_ALIGN.LEFT)
    if subtitle:
        add_textbox(slide, subtitle, 0.4, 0.75, 12.5, 0.4,
                    font_size=14, color=ACCENT_GOLD, align=PP_ALIGN.LEFT, italic=True)
    add_rect(slide, 0.4, 1.18, 12.5, 0.04, fill_color=ACCENT_GOLD)


def add_bullet_block(slide, items, l, t, w, h, font_size=14, color=LIGHT_GRAY, spacing=0.42):
    for i, item in enumerate(items):
        add_textbox(slide, item, l, t + i * spacing, w, spacing,
                    font_size=font_size, color=color)


def add_table(slide, headers, rows, l, t, w, h,
              header_fill=ACCENT_GOLD, row_fills=None):
    cols = len(headers)
    table = slide.shapes.add_table(len(rows)+1, cols,
                                   Inches(l), Inches(t),
                                   Inches(w), Inches(h)).table
    col_w = Inches(w / cols)
    for c in range(cols):
        table.columns[c].width = col_w

    def cell_style(cell, text, fill, txt_color, bold=False, size=12, align=PP_ALIGN.CENTER):
        cell.fill.solid()
        cell.fill.fore_color.rgb = fill
        tf = cell.text_frame
        tf.word_wrap = True
        p = tf.paragraphs[0]
        p.alignment = align
        run = p.add_run()
        run.text = text
        run.font.size  = Pt(size)
        run.font.bold  = bold
        run.font.color.rgb = txt_color

    # header row
    for c, h_text in enumerate(headers):
        cell_style(table.cell(0, c), h_text, header_fill, BG_DARK, bold=True, size=11)

    alt = RGBColor(0x1A, 0x34, 0x52)
    for r, row in enumerate(rows):
        fill = alt if r % 2 == 0 else BG_SLIDE
        if row_fills and r < len(row_fills) and row_fills[r]:
            fill = row_fills[r]
        for c, val in enumerate(row):
            cell_style(table.cell(r+1, c), val, fill, WHITE, size=10, align=PP_ALIGN.LEFT)


# ════════════════════════════════════════════════════════════════════
# SLIDE 1 — PORTADA
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s, BG_DARK)

# decorative left strip
add_rect(s, 0, 0, 0.6, 7.5, fill_color=ACCENT_GOLD)

# big logo / name
add_textbox(s, "ASESAMO", 1.0, 1.6, 11, 1.5,
            font_size=72, bold=True, color=WHITE, align=PP_ALIGN.LEFT)

# tagline
add_textbox(s, "Seguridad · Higiene Industrial · Cumplimiento",
            1.0, 3.0, 11, 0.5, font_size=20, color=ACCENT_GOLD,
            align=PP_ALIGN.LEFT, italic=True)

add_rect(s, 1.0, 3.55, 10.0, 0.05, fill_color=ACCENT_GOLD)

add_textbox(s,
            "Control de contratistas, REPSE y NOMs en tiempo real\ndesde cualquier celular.",
            1.0, 3.7, 11, 0.9, font_size=18, color=LIGHT_GRAY, align=PP_ALIGN.LEFT)

add_textbox(s, "SaaS B2B  ·  México  ·  Octubre 2026",
            1.0, 6.6, 11, 0.5, font_size=13, color=LIGHT_GRAY,
            align=PP_ALIGN.LEFT, italic=True)


# ════════════════════════════════════════════════════════════════════
# SLIDE 2 — EL PROBLEMA
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "🚨  El Problema",
            "Las plantas industriales en México operan con riesgos legales y operativos críticos sin resolver")

problems = [
    ("1️⃣  Contratista con REPSE vencido",
     "La empresa absorbe el 100% de la deuda del IMSS del contratista",
     ACCENT_RED),
    ("2️⃣  Trabajador sin permiso sufre accidente",
     "Multa STPS: $250,000 – $1,000,000 MXN + paro de operaciones",
     ACCENT_GOLD),
    ("3️⃣  Registro de herramientas en papel",
     "15–20 min por contratista. Filas, errores y robos sin evidencia",
     MID_BLUE),
]
for i, (title, desc, color) in enumerate(problems):
    y = 1.5 + i * 1.5
    add_rect(s, 0.4, y, 12.5, 1.2, fill_color=RGBColor(0x1A, 0x30, 0x4A))
    add_rect(s, 0.4, y, 0.12, 1.2, fill_color=color)
    add_textbox(s, title, 0.65, y + 0.1, 12.0, 0.45,
                font_size=15, bold=True, color=WHITE)
    add_textbox(s, desc,  0.65, y + 0.55, 12.0, 0.45,
                font_size=13, color=LIGHT_GRAY)

add_rect(s, 0.4, 6.0, 12.5, 0.65, fill_color=RGBColor(0x1E, 0x0A, 0x0A))
add_textbox(s,
            "⚡  El 95% de las plantas en México tiene al menos 2 omisiones graves en los expedientes de sus contratistas.",
            0.6, 6.05, 12.2, 0.55, font_size=13, bold=True, color=ACCENT_GOLD)


# ════════════════════════════════════════════════════════════════════
# SLIDE 3 — LA SOLUCIÓN
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "💡  La Solución",
            "Una plataforma que conecta el expediente legal del contratista con la caseta en tiempo real")

steps = [
    ("📱  CONTRATISTA", "Sube DC-3, REPSE, SUA e INE desde su celular antes de llegar a planta", ACCENT_GREEN),
    ("🚪  CASETA", "Guardia escanea QR.\n🟢 VERDE = Pasa\n🔴 ROJO = Bloqueado automáticamente", ACCENT_GOLD),
    ("📊  DIRECCIÓN", "Dashboard en tiempo real:\n• Quién está adentro\n• Docs por vencer\n• Exposición legal $$$", MID_BLUE),
]
for i, (title, desc, color) in enumerate(steps):
    x = 0.4 + i * 4.3
    add_rect(s, x, 1.45, 3.9, 3.8, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 1.45, 3.9, 0.12, fill_color=color)
    add_textbox(s, title, x + 0.15, 1.6, 3.6, 0.55,
                font_size=16, bold=True, color=color)
    add_textbox(s, desc, x + 0.15, 2.2, 3.6, 2.8,
                font_size=13, color=LIGHT_GRAY)
    if i < 2:
        add_textbox(s, "→", x + 4.05, 3.1, 0.3, 0.5,
                    font_size=26, bold=True, color=ACCENT_GOLD, align=PP_ALIGN.CENTER)

bullets = [
    "✅  Check-in en 5 segundos con código QR desde cualquier celular",
    "✅  Valida DC-3, REPSE, SUA, INE y exámenes médicos automáticamente",
    "✅  Bloquea el acceso si algún documento está vencido",
    "✅  Alertas automáticas 7 días antes del vencimiento",
]
for i, b in enumerate(bullets):
    add_textbox(s, b, 0.4, 5.5 + i * 0.38, 12.5, 0.35,
                font_size=12, color=LIGHT_GRAY)


# ════════════════════════════════════════════════════════════════════
# SLIDE 4 — MERCADO
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "🎯  Mercado Objetivo",
            "~59,500 plantas industriales en México. TAM: $2,130M MXN/año")

add_table(s,
    ["Plan", "Plantas estimadas", "Ticket/mes", "Mercado Anual"],
    [
        ["🟢  Starter  (PyME)", "~45,000 plantas", "$1,999 MXN", "$1,079M MXN"],
        ["🔵  Industrial Pro",  "~12,000 plantas", "$4,800 MXN",  "$691M MXN"],
        ["🔴  Enterprise",      "~2,500 plantas",  "$12,000 MXN", "$360M MXN"],
        ["TOTAL  TAM",          "~59,500 plantas",  "—",          "$2,130M MXN"],
    ],
    l=0.4, t=1.5, w=12.5, h=2.2,
)

add_textbox(s, "📍  Geografía Prioritaria — Fase 1", 0.4, 3.9, 12.5, 0.45,
            font_size=15, bold=True, color=ACCENT_GOLD)

zones = [
    ("NORTE", "Monterrey · Saltillo\nTijuana · Juárez"),
    ("BAJÍO", "Querétaro · Guanajuato\nSan Luis Potosí · León"),
    ("CENTRO", "Edomex · Puebla · CDMX"),
]
for i, (region, cities) in enumerate(zones):
    x = 0.4 + i * 4.3
    add_rect(s, x, 4.4, 3.9, 1.6, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_textbox(s, region, x + 0.1, 4.45, 3.7, 0.4,
                font_size=14, bold=True, color=ACCENT_GOLD)
    add_textbox(s, cities, x + 0.1, 4.85, 3.7, 1.0,
                font_size=12, color=LIGHT_GRAY)

add_textbox(s,
            "🔥  Industrias calientes: Alimentos y Bebidas · Automotriz · Logística / CEDIS · Construcción Industrial",
            0.4, 6.15, 12.5, 0.5, font_size=12, color=LIGHT_GRAY, italic=True)


# ════════════════════════════════════════════════════════════════════
# SLIDE 5 — COMPETENCIA
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "🏆  Panorama Competitivo",
            "El mercado tiene una brecha enorme: nadie atiende a la PyME industrial con una solución integral")

add_table(s,
    ["Competidor", "Precio", "¿Resuelve campo?", "¿Accesible PyME?"],
    [
        ["Twind / CTAIMA",        "Alto (enterprise)", "Parcialmente",  "❌  No"],
        ["Clave 10",              "Alto",              "Sí",            "❌  No"],
        ["ISOTools / HSETools",   "Alto",              "Parcialmente",  "❌  No"],
        ["Dokify / e-coordina",   "Muy alto",          "Parcialmente",  "❌  No"],
        ["Verifty / FAMA REPSE",  "Medio",             "❌  Solo fiscal","❌  No"],
        ["SafetyCulture",         "Medio",             "❌  Solo campo", "Parcialmente"],
        ["★  ASESAMO",            "Bajo – Medio",      "✅  Integral",  "✅  Sí"],
    ],
    l=0.4, t=1.5, w=12.5, h=3.5,
    row_fills=[None,None,None,None,None,None, RGBColor(0x0D, 0x3D, 0x1C)],
)

add_rect(s, 0.4, 5.2, 12.5, 0.9, fill_color=RGBColor(0x0D, 0x3D, 0x1C))
add_textbox(s,
            "🎯  Nuestra ventaja: Único en combinar precio accesible para PyME + integración operativa-legal completa + IA para documentos",
            0.6, 5.28, 12.2, 0.75, font_size=13, bold=True, color=ACCENT_GREEN)


# ════════════════════════════════════════════════════════════════════
# SLIDE 6 — PRODUCTO Y ROADMAP
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "📦  Producto & Roadmap",
            "MVP enfocado → crecimiento modular → diferenciación con IA")

phases = [
    ("FASE 1  —  MVP\nMes 1 – 5",
     "① Check-in QR en caseta\n② Portal del Contratista\n③ Semáforo Documental\n④ Dashboard Gerente EHS",
     "Lo que genera los primeros ingresos", ACCENT_GREEN),
    ("FASE 2  —  CRECIMIENTO\nMes 6 – 10",
     "⑤ Inducción Express (LMS)\n⑥ PTAR Digital (alto riesgo)\n⑦ Reporte de Incidentes\n⑧ Checklist EPP móvil",
     "Sube el ARPU y reduce el churn", ACCENT_GOLD),
    ("FASE 3  —  IA PREMIUM\nMes 11 – 18",
     "⑨ OCR Inteligente de docs\n⑩ Visión Artificial de EPP\n⑪ EHS Copilot (asesor IA)\n⑫ LOTO Digital",
     "Diferenciación Enterprise imposible de copiar", MID_BLUE),
]
for i, (phase, items, note, color) in enumerate(phases):
    x = 0.4 + i * 4.3
    add_rect(s, x, 1.45, 3.9, 4.5, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 1.45, 3.9, 0.1, fill_color=color)
    add_textbox(s, phase, x + 0.12, 1.6, 3.68, 0.7,
                font_size=13, bold=True, color=color)
    add_textbox(s, items, x + 0.12, 2.38, 3.68, 2.5,
                font_size=12, color=LIGHT_GRAY)
    add_rect(s, x + 0.12, 4.88, 3.66, 0.04, fill_color=color)
    add_textbox(s, note, x + 0.12, 4.95, 3.68, 0.75,
                font_size=11, color=color, italic=True)


# ════════════════════════════════════════════════════════════════════
# SLIDE 7 — MODELO DE NEGOCIO
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "💰  Modelo de Negocio",
            "Suscripción mensual recurrente (SaaS) con múltiples fuentes de monetización")

add_table(s,
    ["Plan", "Precio / mes", "Contratistas", "Incluye"],
    [
        ["🟢  Starter (PyME)",    "$1,999 MXN", "Hasta 25",     "QR check-in · Portal · Alertas"],
        ["🔵  Industrial Pro",    "$4,800 MXN", "Hasta 150",    "+ PTAR · LMS · Incidentes"],
        ["🔴  Enterprise",        "$9,500 – $16,000 MXN", "Ilimitados", "+ APIs torniquetes · IA · Integraciones ERP"],
    ],
    l=0.4, t=1.5, w=12.5, h=1.8,
)

add_textbox(s, "Fuentes adicionales de ingresos", 0.4, 3.5, 12.5, 0.45,
            font_size=15, bold=True, color=ACCENT_GOLD)

extra = [
    ("💼  Modelo Paga el Contratista",
     "$150 – $350 MXN por trabajador / año para mantener expediente activo en la plataforma.",
     ACCENT_GOLD),
    ("🤖  Módulo IA Premium (Add-on)",
     "OCR + Visión EPP: $500 – $1,500 MXN/mes adicionales sobre cualquier plan.",
     MID_BLUE),
    ("🤝  Canal Aliados (comisión 17.5%)",
     "Recurrente de por vida para consultores EHS y despachos laborales que refieran clientes.",
     ACCENT_GREEN),
]
for i, (title, desc, color) in enumerate(extra):
    y = 4.05 + i * 0.95
    add_rect(s, 0.4, y, 12.5, 0.82, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, 0.4, y, 0.08, 0.82, fill_color=color)
    add_textbox(s, title, 0.6, y + 0.05, 4.5, 0.35, font_size=13, bold=True, color=color)
    add_textbox(s, desc,  0.6, y + 0.42, 12.0, 0.35, font_size=12, color=LIGHT_GRAY)


# ════════════════════════════════════════════════════════════════════
# SLIDE 8 — UNIT ECONOMICS
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "📊  Métricas Unitarias",
            "Un modelo de negocio altamente eficiente en capital")

metrics = [
    ("$4,540 MXN",  "ARPA\n(ingreso por cliente / mes)",  ACCENT_GOLD),
    ("1.8%",        "Churn mensual\nesperado",             MID_BLUE),
    ("$252,222 MXN","LTV\n(escenario base)",               ACCENT_GREEN),
    ("$3,800 MXN",  "CAC blended\n(costo por cliente)",    ACCENT_GOLD),
]
for i, (val, label, color) in enumerate(metrics):
    x = 0.4 + i * 3.2
    add_rect(s, x, 1.45, 3.0, 2.0, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 1.45, 3.0, 0.1, fill_color=color)
    add_textbox(s, val,   x + 0.15, 1.7,  2.7, 0.8,
                font_size=26, bold=True, color=color, align=PP_ALIGN.CENTER)
    add_textbox(s, label, x + 0.15, 2.5,  2.7, 0.8,
                font_size=12, color=LIGHT_GRAY, align=PP_ALIGN.CENTER)

# LTV/CAC big call-out
add_rect(s, 0.4, 3.65, 12.5, 1.35, fill_color=RGBColor(0x0D, 0x3D, 0x1C))
add_textbox(s, "LTV / CAC  =  66×", 0.6, 3.75, 12.2, 0.6,
            font_size=34, bold=True, color=ACCENT_GREEN, align=PP_ALIGN.CENTER)
add_textbox(s, "Benchmark saludable para un SaaS: ≥ 3×   |   Payback del CAC: < 2 meses   |   Margen bruto real: 43%–50%",
            0.6, 4.3, 12.2, 0.45, font_size=12, color=LIGHT_GRAY, align=PP_ALIGN.CENTER)

add_table(s,
    ["Clientes de pago", "MRR", "¿Cubre costos?"],
    [
        ["10 clientes",  "$45,400 MXN",  "❌  No cubre"],
        ["17 clientes",  "$77,180 MXN",  "✅  Break-even Fase 1"],
        ["25 clientes",  "$113,500 MXN", "✅  Break-even Fase 2"],
        ["58 clientes",  "$263,320 MXN", "✅  Meta Año 1"],
    ],
    l=0.4, t=5.2, w=12.5, h=1.8,
    row_fills=[None, None, RGBColor(0x0D,0x3D,0x1C), RGBColor(0x0D,0x3D,0x1C)],
)


# ════════════════════════════════════════════════════════════════════
# SLIDE 9 — PROYECCIÓN FINANCIERA
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "📈  Proyección Financiera",
            "Escenario base — primeros 18 meses")

add_table(s,
    ["Indicador", "🟢 Optimista", "🟡 Base", "🔴 Pesimista"],
    [
        ["Clientes mes 12",        "95",           "58",          "22"],
        ["MRR mes 12",             "$431,300 MXN", "$263,320 MXN","$99,880 MXN"],
        ["ARR año 1",              "$5.2M MXN",    "$3.2M MXN",   "$1.2M MXN"],
        ["Break-even mensual",     "Mes 4–5",      "Mes 6",       "Mes 13–14"],
        ["Recuperación inversión", "Mes 9",        "Mes 12",      "Mes 20+"],
        ["Capital requerido",      "$280K MXN",    "$500K MXN",   "$800K+ MXN"],
    ],
    l=0.4, t=1.5, w=12.5, h=2.8,
)

add_textbox(s, "ARR Proyectado por Año", 0.4, 4.5, 12.5, 0.45,
            font_size=15, bold=True, color=ACCENT_GOLD)

years = [
    ("Año 1", "58 clientes", "$3.2M MXN", "~$160K USD", ACCENT_GREEN),
    ("Año 2", "180 clientes","$9.8M MXN", "~$490K USD", ACCENT_GOLD),
    ("Año 3", "400 clientes","$21.8M MXN","~$1.09M USD",MID_BLUE),
]
for i, (yr, clients, arr_mxn, arr_usd, color) in enumerate(years):
    x = 0.4 + i * 4.3
    add_rect(s, x, 5.05, 3.9, 1.8, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 5.05, 3.9, 0.1, fill_color=color)
    add_textbox(s, yr,       x+0.15, 5.2,  3.6, 0.45, font_size=15, bold=True, color=color)
    add_textbox(s, clients,  x+0.15, 5.65, 3.6, 0.35, font_size=12, color=LIGHT_GRAY)
    add_textbox(s, arr_mxn, x+0.15, 6.0,  3.6, 0.4,  font_size=14, bold=True, color=WHITE)
    add_textbox(s, arr_usd, x+0.15, 6.4,  3.6, 0.35, font_size=11, color=LIGHT_GRAY, italic=True)


# ════════════════════════════════════════════════════════════════════
# SLIDE 10 — ESTRATEGIA DE VENTAS
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "🛡️  Estrategia de Ventas B2B",
            "Proceso de cierre en 2–4 semanas con oferta irresistible de entrada")

steps_sales = [
    ("PASO 1\nDemo 20 min", "Mostramos multas STPS, caseta saturada y app móvil en vivo", ACCENT_GOLD),
    ("PASO 2\nPiloto 14 días", "1 caseta real, 2 contratistas reales. Sin costo, sin riesgo", ACCENT_GREEN),
    ("PASO 3\nRevisión con Dirección", "Presentamos: 3 DC-3 vencidos detectados.\nTiempo de ingreso -70%", MID_BLUE),
    ("PASO 4\nContrato Anual", "Cierre. Opción anualidad: 2 meses gratis = -16% sobre precio de lista", ACCENT_GOLD),
]
for i, (step, desc, color) in enumerate(steps_sales):
    x = 0.4 + i * 3.2
    add_rect(s, x, 1.45, 3.0, 2.4, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 1.45, 3.0, 0.12, fill_color=color)
    add_textbox(s, step, x+0.12, 1.62, 2.78, 0.75, font_size=13, bold=True, color=color)
    add_textbox(s, desc, x+0.12, 2.4,  2.78, 1.2,  font_size=12, color=LIGHT_GRAY)
    if i < 3:
        add_textbox(s, "→", x+3.1, 2.35, 0.2, 0.5,
                    font_size=22, bold=True, color=ACCENT_GOLD, align=PP_ALIGN.CENTER)

add_rect(s, 0.4, 4.05, 12.5, 1.7, fill_color=RGBColor(0x1A, 0x2C, 0x1A))
add_textbox(s, "🐴  El Caballo de Troya — Oferta de Entrada Gratuita",
            0.6, 4.12, 12.2, 0.45, font_size=14, bold=True, color=ACCENT_GREEN)
add_textbox(s,
            '"Permítenos auditar de forma confidencial los últimos 5 expedientes de contratistas. '
            'Si encontramos omisiones que te expongan a una multa de la STPS, te entregamos el reporte legal sin costo."\n'
            '→  El 95% de las plantas tiene al menos 2 omisiones graves. Al ver su propio riesgo, la venta queda prácticamente cerrada.',
            0.6, 4.6, 12.2, 1.05, font_size=12, color=LIGHT_GRAY)

add_table(s,
    ["Canal", "CAC", "Escalabilidad"],
    [
        ["🤝  Alianzas (consultores EHS, aseguradoras, despachos)", "Bajo (comisión diferida)", "🟢  Muy alta"],
        ["📱  Outbound LinkedIn + visitas a parques",                "$4,500 MXN",               "🟡  Media"],
        ["📝  Inbound (SEO, webinars, herramientas gratuitas)",      "$2,000 MXN",               "🟢  Alta a LP"],
    ],
    l=0.4, t=5.85, w=12.5, h=1.4,
)


# ════════════════════════════════════════════════════════════════════
# SLIDE 11 — ROADMAP DE EJECUCIÓN
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "🗓️  Roadmap de Ejecución — 18 Meses", "")

add_table(s,
    ["Periodo", "Actividad Principal", "Meta Comercial"],
    [
        ["Mes 1–2",   "Desarrollo MVP (4 módulos core). 2 pilotos gratuitos en casetas reales.", "0 clientes pago · Validación de producto"],
        ["Mes 3",     "Lanzamiento comercial. Diagnósticos REPSE gratuitos. LinkedIn outbound.", "3 clientes de pago"],
        ["Mes 4–5",   "5 aliados EHS activos. Primeros artículos SEO. Demos 2x/semana.",         "12 clientes de pago"],
        ["Mes 6  ★",  "Break-even mensual. Inicio desarrollo Fase 2 (PTAR, LMS).",               "18 clientes · MRR $81K MXN"],
        ["Mes 7–9",   "Módulos Fase 2 en producción. Primer caso de éxito publicado.",            "37 clientes de pago"],
        ["Mes 10–12 ★","Break-even acumulado. Webinar mensual. Canal aliados consolidado.",       "58 clientes · ARR $3.2M MXN"],
        ["Mes 13–18", "Módulos IA (OCR, Visión EPP, Copilot). Expansión segunda ciudad.",        "108 clientes · ARR $5.9M MXN"],
    ],
    l=0.4, t=1.45, w=12.5, h=5.7,
    row_fills=[None, None, None,
               RGBColor(0x0D,0x3D,0x1C),
               None,
               RGBColor(0x0D,0x3D,0x1C),
               None],
)


# ════════════════════════════════════════════════════════════════════
# SLIDE 12 — INVERSIÓN REQUERIDA
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "🧮  Inversión Requerida",
            "Capital necesario para llegar al break-even sin depender de inversión externa")

# Big number
add_rect(s, 0.4, 1.45, 12.5, 1.2, fill_color=RGBColor(0x14, 0x2B, 0x42))
add_textbox(s, "$500,000 MXN  (~$25,000 USD)",
            0.6, 1.52, 12.2, 0.65, font_size=30, bold=True, color=ACCENT_GOLD, align=PP_ALIGN.CENTER)
add_textbox(s, "Capital requerido para 12 meses de operación hasta break-even acumulado",
            0.6, 2.1, 12.2, 0.4, font_size=12, color=LIGHT_GRAY, align=PP_ALIGN.CENTER)

# breakdown bars
allocs = [
    ("Desarrollo del producto", 56, ACCENT_GREEN),
    ("Ventas y prospección",    24, ACCENT_GOLD),
    ("Legal, contabilidad y marca", 12, MID_BLUE),
    ("Contingencia e infraestructura", 8, ACCENT_RED),
]
for i, (label, pct, color) in enumerate(allocs):
    y = 2.75 + i * 0.8
    add_rect(s, 0.4, y, 8.0 * pct / 100, 0.5, fill_color=color)
    add_rect(s, 0.4 + 8.0*pct/100, y, 8.0*(1-pct/100), 0.5,
             fill_color=RGBColor(0x1A, 0x34, 0x52))
    add_textbox(s, f"{pct}%  –  {label}", 8.65, y + 0.08, 4.6, 0.45,
                font_size=12, color=LIGHT_GRAY)

add_table(s,
    ["Hito alcanzado", "Mes"],
    [
        ["MVP en producción + primeros 3 clientes de pago", "Mes 3"],
        ["18 clientes de pago, break-even mensual",          "Mes 6"],
        ["58 clientes, recuperación total de inversión",     "Mes 12"],
        ["108 clientes, ARR $5.9M MXN, módulos IA activos",  "Mes 18"],
    ],
    l=0.4, t=6.0, w=12.5, h=1.35,
)


# ════════════════════════════════════════════════════════════════════
# SLIDE 13 — EQUIPO
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "👥  Equipo Fundador", "")

roles = [
    ("CEO / Fundador", "Ventas · Producto · Visión estratégica",
     "Experiencia en industria / seguridad industrial\no ventas B2B. Conocimiento del dolor del\nmercado desde adentro.", ACCENT_GOLD),
    ("CTO / Co-fundador", "Arquitectura · Go + Next.js · IA",
     "Senior con experiencia en SaaS, APIs y\nplataformas móviles. Dominio de arquitectura\nclean code y despliegue en nube.", MID_BLUE),
    ("Asesor EHS / Legal", "Normativa · NOMs · REPSE · STPS",
     "Consultor certificado en seguridad industrial.\nEx-inspector STPS o auditor ISO 45001.\nRed de contactos en parques industriales.", ACCENT_GREEN),
]
for i, (role, sub, desc, color) in enumerate(roles):
    x = 0.4 + i * 4.3
    add_rect(s, x, 1.45, 3.9, 4.0, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 1.45, 3.9, 0.12, fill_color=color)

    # avatar placeholder circle
    circ = s.shapes.add_shape(9,  # MSO_SHAPE_TYPE.OVAL
                               Inches(x + 1.25), Inches(1.7), Inches(1.4), Inches(1.4))
    circ.fill.solid()
    circ.fill.fore_color.rgb = RGBColor(0x2A, 0x4A, 0x65)
    circ.line.color.rgb = color
    circ.line.width = Pt(2)

    add_textbox(s, "👤", x + 1.55, 1.82, 0.9, 0.9,
                font_size=28, align=PP_ALIGN.CENTER)
    add_textbox(s, role, x + 0.15, 3.25, 3.6, 0.5,
                font_size=14, bold=True, color=color, align=PP_ALIGN.CENTER)
    add_textbox(s, sub,  x + 0.15, 3.7,  3.6, 0.4,
                font_size=11, color=ACCENT_GOLD, align=PP_ALIGN.CENTER, italic=True)
    add_textbox(s, desc, x + 0.15, 4.2,  3.6, 1.1,
                font_size=11, color=LIGHT_GRAY, align=PP_ALIGN.CENTER)

add_rect(s, 0.4, 5.65, 12.5, 0.9, fill_color=RGBColor(0x14, 0x2B, 0x42))
add_textbox(s, "Advisors estratégicos recomendados:",
            0.6, 5.7, 12.0, 0.35, font_size=13, bold=True, color=ACCENT_GOLD)
add_textbox(s,
            "Corredor de seguros industriales  ·  Ex-director de parque industrial  ·  Abogado laboral especialista en subcontratación",
            0.6, 6.05, 12.0, 0.4, font_size=12, color=LIGHT_GRAY)


# ════════════════════════════════════════════════════════════════════
# SLIDE 14 — POR QUÉ AHORA
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s)
header_bar(s)
slide_title(s, "⏰  ¿Por Qué Ahora?",
            "El mercado está formado. El segmento PyME sigue sin solución. El costo de IA bajó 95%.")

timeline = [
    ("2021", "Reforma de subcontratación\n(REPSE obligatorio)",
     "Las plantas DEBEN validar a sus contratistas por ley.\nAntes de 2021 este negocio no existía.", ACCENT_RED),
    ("2022–25", "STPS intensifica inspecciones.\nMultas históricamente altas.",
     "Los grandes competidores siguen ignorando\nal segmento PyME industrial.", ACCENT_GOLD),
    ("2026 ★", "Punto de inflexión.\nIA accesible. Mercado formado.",
     "ASESAMO entra cuando el dolor es máximo\ny la solución más barata que nunca.", ACCENT_GREEN),
]
for i, (year, title, desc, color) in enumerate(timeline):
    x = 0.4 + i * 4.3
    add_rect(s, x, 1.45, 3.9, 3.8, fill_color=RGBColor(0x14, 0x2B, 0x42))
    add_rect(s, x, 1.45, 3.9, 0.12, fill_color=color)
    add_textbox(s, year,  x+0.15, 1.62, 3.6, 0.55, font_size=22, bold=True, color=color)
    add_textbox(s, title, x+0.15, 2.25, 3.6, 0.9,  font_size=13, bold=True, color=WHITE)
    add_textbox(s, desc,  x+0.15, 3.2,  3.6, 1.8,  font_size=12, color=LIGHT_GRAY)

tailwinds = [
    "🔵  +1.2M empresas registradas en el REPSE generan demanda permanente de validación documental",
    "🟢  El costo de analizar una foto o PDF con IA bajó un 95% desde 2022 (Gemini Flash: $0.0003 USD/análisis)",
    "🟡  Ningún competidor ha lanzado una versión accesible para PyME con self-service onboarding en México",
    "🔴  Las multas STPS en UMAs alcanzan niveles históricos; el ROI de este software es inmediato y medible",
]
for i, t in enumerate(tailwinds):
    add_textbox(s, t, 0.4, 5.45 + i*0.44, 12.5, 0.4, font_size=12, color=LIGHT_GRAY)


# ════════════════════════════════════════════════════════════════════
# SLIDE 15 — RESUMEN EJECUTIVO / CIERRE
# ════════════════════════════════════════════════════════════════════
s = prs.slides.add_slide(BLANK)
bg(s, BG_DARK)
add_rect(s, 0, 0, 0.6, 7.5, fill_color=ACCENT_GOLD)

add_textbox(s, "RESUMEN EJECUTIVO", 0.9, 0.3, 12.0, 0.55,
            font_size=14, bold=True, color=ACCENT_GOLD, italic=True)
add_textbox(s, "ASESAMO en una página", 0.9, 0.75, 12.0, 0.65,
            font_size=26, bold=True, color=WHITE)
add_rect(s, 0.9, 1.35, 11.6, 0.05, fill_color=ACCENT_GOLD)

rows_summary = [
    ["🚨  Problema",    "Plantas industriales gestionan contratistas con Excel y carpetas, expuestas a multas STPS de hasta $1M MXN"],
    ["💡  Solución",    "SaaS que integra control QR de acceso + validación REPSE automática + dashboard legal en tiempo real"],
    ["🎯  Mercado",     "~59,500 plantas en México. TAM $2,130M MXN/año. Foco inicial: PyME industrial del Bajío y Norte"],
    ["💰  Modelo",      "Suscripción $1,999–$16,000 MXN/mes + canal aliados 17.5% + add-on IA Premium"],
    ["🏆  Ventaja",     "Único en combinar: precio PyME + integración operativa-legal completa + IA para documentos"],
    ["📊  Métricas",    "LTV/CAC 66×  ·  Payback < 2 meses  ·  Margen bruto infra 97%"],
    ["📈  Proyección",  "58 clientes y ARR $3.2M MXN al mes 12  ·  108 clientes y ARR $5.9M MXN al mes 18"],
    ["🧮  Inversión",   "$500,000 MXN (~$25,000 USD) para 12 meses hasta break-even acumulado"],
]
for i, (label, val) in enumerate(rows_summary):
    y = 1.55 + i * 0.63
    fill = RGBColor(0x1A, 0x30, 0x4A) if i % 2 == 0 else RGBColor(0x14, 0x26, 0x38)
    add_rect(s, 0.9, y, 11.6, 0.58, fill_color=fill)
    add_textbox(s, label, 0.95, y+0.08, 2.0, 0.45, font_size=12, bold=True, color=ACCENT_GOLD)
    add_textbox(s, val,   2.95, y+0.08, 9.4, 0.45, font_size=12, color=WHITE)

add_rect(s, 0.9, 6.65, 11.6, 0.7, fill_color=ACCENT_GOLD)
add_textbox(s,
            "No vendemos software. Vendemos la única herramienta que evita que un director de planta reciba una multa de $500K MXN el lunes.",
            0.98, 6.7, 11.45, 0.58, font_size=12, bold=True, color=BG_DARK, align=PP_ALIGN.CENTER)


# ── Save ────────────────────────────────────────────────────────────
out = "/Users/jasbotello/devel/asesamo/ASESAMO - Pitch Deck.pptx"
prs.save(out)
print(f"✅  Saved → {out}")
