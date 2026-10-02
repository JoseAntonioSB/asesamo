# 💰 Complemento Financiero — SAS Seguridad e Higiene ASESAMO

> Este documento completa las brechas identificadas en el plan de negocios original. Debe leerse en conjunto con el documento principal.

---

## ÍNDICE

1. [Inversión Inicial Requerida](#1-inversión-inicial-requerida)
2. [Alcance Real del MVP](#2-alcance-real-del-mvp--hoja-de-ruta-de-producto)
3. [Estructura de Costos Operativos Completa](#3-estructura-de-costos-operativos-completa)
4. [Métricas de Retención — Churn / LTV / CAC](#4-métricas-de-retención--churn--ltv--cac)
5. [Estado de Resultados Mensual — Primeros 18 Meses](#5-estado-de-resultados-mensual--primeros-18-meses-escenario-base)
6. [3 Escenarios Financieros](#6-tres-escenarios-financieros-mes-12)
7. [Riesgo Regulatorio y Mitigación](#7-riesgo-regulatorio-y-mitigación)
8. [Estrategia de Contenido y Marketing Digital](#8-estrategia-de-contenido-y-marketing-digital)
9. [Resumen Ejecutivo Financiero](#9-resumen-ejecutivo-financiero)

---

## 1. Inversión Inicial Requerida

### Supuestos del Equipo (Bootstrapped — Escenario Recomendado)

| Rol | Modalidad | Costo Mensual Estimado (MXN) |
|---|---|---|
| Desarrollador Senior Full-Stack (Go + Next.js) | Freelance / Socio técnico | $60,000 – $80,000 |
| Fundador / CEO (ventas, producto, operación) | Propio capital de trabajo | $0 (reinversión) |
| Asesor Legal Laboral (REPSE / NOMs) | Por proyecto / honorarios | $5,000 – $8,000 |
| Contador / CFO externo | Mensual fijo | $3,500 – $5,000 |

> [!NOTE]
> El escenario más viable para arrancar es que el fundador sea el responsable de ventas y el desarrollador sea co-fundador con participación accionaria (equity) en lugar de sueldo completo, reduciendo el burn rate mensual significativamente.

---

### Capital Necesario para Llegar al Break-Even

El break-even ocurre aproximadamente en el **mes 9** (ver sección 5). Por lo tanto, se necesita financiar los primeros ~9 meses de operación antes de que el negocio se sostenga solo.

| Concepto | Meses 1–6 (Desarrollo + Primeros Clientes) | Meses 7–9 (Pre Break-Even) | Total |
|---|---|---|---|
| Desarrollo (desarrollador) | $420,000 MXN | $210,000 MXN | $630,000 MXN |
| Legal + Contabilidad | $48,000 MXN | $24,000 MXN | $72,000 MXN |
| Infraestructura y herramientas | $5,000 MXN | $12,000 MXN | $17,000 MXN |
| Ventas y prospección (LinkedIn, viáticos) | $30,000 MXN | $30,000 MXN | $60,000 MXN |
| Dominio, marca, diseño inicial | $15,000 MXN | $0 | $15,000 MXN |
| Colchón de contingencia (10%) | $51,800 MXN | $27,600 MXN | $79,400 MXN |
| **TOTAL** | **$569,800 MXN** | **$303,600 MXN** | **$873,400 MXN** |
| Menos ingresos esperados en ese periodo | -$0 (M1–M2) / -$110,000 (M3–M6) | -$280,000 MXN | **-$390,000 MXN** |
| **Capital Neto Requerido** | | | **≈ $480,000 – $520,000 MXN** |

> [!IMPORTANT]
> **Inversión inicial mínima recomendada: $500,000 MXN (~$25,000 USD)**
> 
> Este capital puede provenir de: ahorros propios, préstamo personal, socio inversionista ángel, o aceleradoras como Startup Mexico, NAFIN Emprendedor o fondos CONACYT-INADEM.

---

### Alternativa: Co-fundador Técnico con Equity

Si el desarrollador acepta equity (participación del 15–25%) en lugar de sueldo completo, el capital requerido se reduce a:

| Concepto | Estimado |
|---|---|
| Capital mínimo de arranque | **$150,000 – $200,000 MXN** |
| Equity cedido al co-fundador técnico | 15% – 25% de la empresa |

---

## 2. Alcance Real del MVP — Hoja de Ruta de Producto

### Principio: "Lanzar algo pequeño que funcione perfectamente" > "Lanzar algo grande que funcione a medias"

El plan original propone 10+ módulos para el MVP. Aquí está la priorización correcta basada en impacto vs. esfuerzo:

### Fase MVP — Meses 1 a 5 (El Producto Mínimo que te Genera Ingresos)

| # | Módulo | Descripción | Impacto de Venta | Complejidad |
|---|---|---|---|---|
| 1 | **Check-in QR en Caseta** | El guardia escanea el QR del contratista. Pantalla verde/rojo en 2 segundos. | 🔴 Crítico | Baja |
| 2 | **Portal del Contratista** | El contratista crea su expediente, sube DC-3, REPSE, SUA e INE. | 🔴 Crítico | Media |
| 3 | **Validación y Semáforo Documental** | El sistema revisa automáticamente fechas de vigencia y emite alertas. | 🔴 Crítico | Media |
| 4 | **Dashboard de la Planta (Gerente EHS)** | Vista de quién está adentro, qué documentos vencen esta semana, alertas activas. | 🔴 Crítico | Media |

**Tiempo estimado de desarrollo del MVP real:** 10 – 14 semanas  
**Costo estimado (desarrollador freelance senior):** $150,000 – $200,000 MXN

---

### Fase 2 — Meses 6 a 10 (Módulos que Aumentan el ARPU y Reducen el Churn)

| # | Módulo | Por qué en Fase 2 |
|---|---|---|
| 5 | **Inducción Express (LMS básico)** | Aumenta el valor percibido y reduce congestión en caseta. |
| 6 | **PTAR Digital (Permisos de Alto Riesgo)** | Sube al Plan Industrial Pro; justifica precio más alto. |
| 7 | **Módulo de Incidentes / Near Misses** | Engancha al Gerente EHS; datos clave para renovar contrato. |
| 8 | **Inspección de EPP con Checklist Móvil** | Fácil de construir sobre el check-in QR ya existente. |

---

### Fase 3 — Meses 11 a 18 (Módulos Premium con IA — Diferenciador Enterprise)

| # | Módulo | Por qué en Fase 3 |
|---|---|---|
| 9 | **OCR Inteligente de Documentos (Document AI)** | Requiere afinar prompts y validar con volumen real de documentos. |
| 10 | **Detección de EPP por Visión Artificial** | Requiere reentrenamiento con fotos reales de las plantas clientes. |
| 11 | **Generador de AST con IA (EHS Copilot)** | Funciona mejor cuando hay historial de incidentes en el sistema. |
| 12 | **LOTO Digital / Módulo de Energías Peligrosas** | Alta complejidad regulatoria; implementar con primeros clientes Enterprise. |
| 13 | **Evacuación y Mustering en Tiempo Real** | Se construye sobre la base del control de accesos ya operando. |

---

## 3. Estructura de Costos Operativos Completa

A diferencia del plan original (que solo incluye costos de infraestructura), esta es la estructura **real** de costos mes a mes:

### Fase 1: MVP + Primeros Clientes (Meses 1–6)

| Concepto | Proveedor | Costo Mensual (MXN) |
|---|---|---|
| **Desarrollo (desarrollador senior)** | Freelance / Co-fundador | $60,000 – $70,000 |
| **Infraestructura (servidor, R2, IA)** | DigitalOcean + Cloudflare | $380 |
| **Asesor Legal Laboral** | Honorarios mensuales | $6,000 |
| **Contabilidad y Facturación** | Despacho externo | $4,000 |
| **LinkedIn Sales Navigator** | Prospección outbound | $1,750 |
| **Viáticos / Visitas a Parques** | Combustible, comidas | $3,000 |
| **Herramientas SaaS (Figma, Notion, Slack, GitHub)** | Varias | $1,500 |
| **Dominio + Email corporativo (Google Workspace)** | Cloudflare + Google | $500 |
| **TOTAL COSTOS FIJOS FASE 1** | | **$77,130 MXN/mes** |

---

### Fase 2: Tracción (Meses 7–12, con 10–30 clientes)

| Concepto | Proveedor | Costo Mensual (MXN) |
|---|---|---|
| **Desarrollo + Soporte técnico** | Desarrollador (aumento o segundo) | $75,000 |
| **Infraestructura (Fase 2)** | DigitalOcean + Cloudflare R2 | $3,500 |
| **Comisiones a Aliados (canal referidos)** | 17.5% promedio sobre MRR de esos clientes | Variable (~$8,000–$18,000) |
| **Atención a Clientes / Customer Success** | Tiempo del fundador o asistente part-time | $8,000 |
| **Legal + Contabilidad** | Mayor volumen de contratos y facturas | $9,000 |
| **Marketing (LinkedIn Ads + Contenido)** | Automatización y anuncios focalizados | $6,000 |
| **TOTAL COSTOS FIJOS + VARIABLES FASE 2** | | **$109,500 – $119,500 MXN/mes** |

---

## 4. Métricas de Retención — Churn / LTV / CAC

### Supuestos de Precios y Mix de Planes

| Plan | Precio MXN/mes | Mix esperado de clientes |
|---|---|---|
| Starter (PyME) | $1,999 | 35% |
| Industrial Pro | $4,800 | 55% |
| Enterprise | $12,000 (promedio) | 10% |

**ARPA (Average Revenue Per Account):**  
`(0.35 × 1,999) + (0.55 × 4,800) + (0.10 × 12,000) = $699.65 + $2,640 + $1,200 = $4,540 MXN/mes`

---

### Churn Rate — Tasa de Cancelación Mensual

El churn en SaaS B2B industrial es **naturalmente bajo** porque:
- El costo de cambiar de plataforma es alto (re-capacitar a guardias, migrar expedientes).
- El sistema se vuelve un requisito legal interno de la planta.
- Los contratos anuales eliminan el churn mensual el primer año.

| Escenario | Churn Mensual | Justificación |
|---|---|---|
| Optimista | 1.0% | Contratos anuales, alta dependencia operativa |
| Base (Recomendado) | 1.8% | Algunas PyMEs con ajustes de presupuesto |
| Pesimista | 3.5% | Alta rotación si el producto no entrega valor rápido |

---

### LTV — Lifetime Value por Cliente

**Fórmula:** `LTV = ARPA / Churn Mensual`

| Escenario | ARPA | Churn | LTV |
|---|---|---|---|
| Optimista | $4,540 MXN | 1.0% | **$454,000 MXN** (~$22,700 USD) |
| Base | $4,540 MXN | 1.8% | **$252,222 MXN** (~$12,611 USD) |
| Pesimista | $4,540 MXN | 3.5% | **$129,714 MXN** (~$6,486 USD) |

---

### CAC — Costo de Adquisición por Cliente

| Canal | Costo Estimado por Cliente Cerrado | Supuesto |
|---|---|---|
| Outbound directo (fundador) | $4,500 MXN | 3 horas/semana prospección, 15 demos para 1 cierre |
| Aliados (consultores EHS) | $0 CAC fijo + 17.5% MRR recurrente | Comisión de por vida = CAC diferido en el LTV |
| Inbound (contenido SEO/LinkedIn) | $2,000 MXN | Bajo volumen inicial pero crece exponencialmente |
| **Blended CAC (promedio ponderado)** | **$3,800 MXN** | Mix 50% outbound / 40% aliados / 10% inbound |

> [!TIP]
> El canal de aliados no tiene CAC inicial pero **reduce el LTV efectivo** en 17.5%. Calcula el LTV ajustado por canal:
> - LTV canal directo (Base): $252,222 MXN
> - LTV canal aliados (Base): $252,222 × (1 - 0.175) = **$207,983 MXN**

---

### Ratio LTV / CAC — El Indicador de Salud del SaaS

**Referencia:** Un SaaS saludable tiene LTV/CAC ≥ 3×. Los mejores SaaS B2B tienen 10×–20×.

| Escenario | LTV | CAC Blended | LTV/CAC | Evaluación |
|---|---|---|---|---|
| Optimista | $454,000 MXN | $3,800 MXN | **119×** | 🟢 Extraordinario |
| Base | $252,222 MXN | $3,800 MXN | **66×** | 🟢 Excelente |
| Pesimista | $129,714 MXN | $3,800 MXN | **34×** | 🟢 Muy saludable |

> [!NOTE]
> El ratio LTV/CAC tan alto se debe al bajo costo de adquisición por prospección directa en parques industriales vs. el ciclo de vida largo de un cliente industrial. Esto confirma que el negocio es **altamente eficiente en capital** una vez que se tiene tracción.

---

### Período de Recuperación del CAC (Payback Period)

**Fórmula:** `Payback = CAC / (ARPA × Margen Bruto)`

Usando margen bruto de infraestructura del 97%:  
`Payback = 3,800 / (4,540 × 0.97) = 3,800 / 4,404 = **0.86 meses**`

Considerando margen bruto **real** (incluyendo todos los costos operativos) de ~45%:  
`Payback = 3,800 / (4,540 × 0.45) = 3,800 / 2,043 = **1.86 meses**`

**El cliente recupera su costo de adquisición en menos de 2 meses.** Esto es excepcional.

---

## 5. Estado de Resultados Mensual — Primeros 18 Meses (Escenario Base)

### Supuestos del Escenario Base:
- Churn mensual: 1.8%
- ARPA: $4,540 MXN/mes
- Costos operativos Fase 1: $77,130 MXN/mes | Fase 2: $114,500 MXN/mes | Fase 3: $135,000 MXN/mes
- Clientes piloto meses 1–2: gratuitos (0 ingresos)
- Crecimiento neto de clientes: conservador, 3–6 nuevos/mes

| Mes | Clientes Pago | MRR (MXN) | Costos Totales (MXN) | Resultado Neto (MXN) | Acumulado (MXN) |
|---|---|---|---|---|---|
| 1 | 0 | $0 | $77,130 | **-$77,130** | -$77,130 |
| 2 | 0 | $0 | $77,130 | **-$77,130** | -$154,260 |
| 3 | 3 | $13,620 | $77,130 | **-$63,510** | -$217,770 |
| 4 | 7 | $31,780 | $77,130 | **-$45,350** | -$263,120 |
| 5 | 12 | $54,480 | $77,130 | **-$22,650** | -$285,770 |
| 6 | 18 | $81,720 | $77,130 | **+$4,590** ✅ | -$281,180 |
| 7 | 24 | $108,960 | $114,500 | **-$5,540** | -$286,720 |
| 8 | 30 | $136,200 | $114,500 | **+$21,700** ✅ | -$265,020 |
| 9 | 37 | $167,980 | $114,500 | **+$53,480** ✅ | -$211,540 |
| 10 | 44 | $199,760 | $114,500 | **+$85,260** ✅ | -$126,280 |
| 11 | 51 | $231,540 | $114,500 | **+$117,040** ✅ | -$9,240 |
| 12 | 58 | $263,320 | $114,500 | **+$148,820** ✅ | **+$139,580** 🎯 |
| 13 | 66 | $299,640 | $135,000 | **+$164,640** ✅ | +$304,220 |
| 14 | 74 | $335,960 | $135,000 | **+$200,960** ✅ | +$505,180 |
| 15 | 83 | $376,820 | $135,000 | **+$241,820** ✅ | +$747,000 |
| 16 | 92 | $417,680 | $135,000 | **+$282,680** ✅ | +$1,029,680 |
| 17 | 100 | $454,000 | $135,000 | **+$319,000** ✅ | +$1,348,680 |
| 18 | 108 | $490,320 | $135,000 | **+$355,320** ✅ | +$1,703,000 |

> [!IMPORTANT]
> **Puntos clave de la proyección base:**
> - 🔴 **Break-even mensual:** Mes 6 (primer mes positivo, aunque marginal)
> - 🔴 **Break-even acumulado (recuperación de inversión):** Mes 12
> - 💰 **ARR proyectado al mes 18:** ~$5,883,840 MXN (~$294,000 USD)
> - 📈 **Clientes necesarios para cubrir costos Fase 2:** ~25 clientes de pago

---

## 6. Tres Escenarios Financieros (Mes 12)

### Diferencias entre escenarios:
- **Optimista:** 6–8 clientes nuevos/mes, churn 1%, alianzas activas desde mes 3
- **Base:** 3–6 clientes nuevos/mes, churn 1.8%, alianzas activas desde mes 5
- **Pesimista:** 1–3 clientes nuevos/mes, churn 3.5%, ciclos de venta largos (6–8 semanas)

| Indicador | 🟢 Optimista | 🟡 Base | 🔴 Pesimista |
|---|---|---|---|
| **Clientes al mes 12** | 95 | 58 | 22 |
| **MRR mes 12** | $431,300 MXN | $263,320 MXN | $99,880 MXN |
| **ARR proyectado** | $5,175,600 MXN | $3,159,840 MXN | $1,198,560 MXN |
| **Resultado neto mes 12** | +$316,800 MXN | +$148,820 MXN | -$14,620 MXN |
| **Break-even mensual** | Mes 4–5 | Mes 6 | Mes 13–14 |
| **Recuperación inversión** | Mes 9 | Mes 12 | Mes 20+ |
| **Capital total necesario** | $280,000 MXN | $500,000 MXN | $800,000+ MXN |

> [!WARNING]
> En el escenario pesimista, el negocio no alcanza el break-even en el primer año. La variable crítica no es el precio ni la tecnología, sino **la velocidad de adquisición de clientes**. El canal de alianzas con consultores EHS es el mayor acelerador del escenario optimista.

---

### Análisis de Sensibilidad — ¿Cuántos Clientes Necesitas para Sobrevivir?

| Clientes de pago activos | MRR Estimado | ¿Cubre costos Fase 1? | ¿Cubre costos Fase 2? |
|---|---|---|---|
| 5 | $22,700 MXN | ❌ No | ❌ No |
| 10 | $45,400 MXN | ❌ No | ❌ No |
| 17 | $77,180 MXN | ✅ Sí (break-even F1) | ❌ No |
| 25 | $113,500 MXN | ✅ Sí | ✅ Sí (break-even F2) |
| 40 | $181,600 MXN | ✅ Sí | ✅ Sí + excedente |

**Conclusión crítica: Con 25 clientes de pago, el negocio es autosustentable. Esta es la meta mínima del año 1.**

---

## 7. Riesgo Regulatorio y Mitigación

El modelo de negocio depende directamente del marco legal del REPSE y las NOMs mexicanas. A continuación, los riesgos regulatorios y cómo mitigarlos:

### Mapa de Riesgos Regulatorios

| Riesgo | Probabilidad | Impacto | Estrategia de Mitigación |
|---|---|---|---|
| **Modificación del régimen REPSE** (nuevas reformas fiscales que cambien los requisitos de subcontratación) | Media | Alto | Diseñar el módulo de cumplimiento como configurable; el sistema se actualiza cambiando reglas en base de datos, no código. |
| **Actualización de NOMs** (la STPS actualiza o deroga una NOM y el checklist del sistema queda desactualizado) | Alta (ocurre ~2–3 veces/año) | Medio | Contratar un asesor EHS externo de forma recurrente (ya en costos). Suscripción a DOF (Diario Oficial de la Federación) con alertas automáticas. |
| **Cambios en la API del IMSS / SAT** (los sistemas de consulta de opinión de cumplimiento cambian su formato) | Media | Alto | Usar un conector intermedio (adapter pattern en Go) que aísle la lógica de validación. Monitorear el portal del IMSS/SAT mensualmente. |
| **Entrada de un gran competidor** (SAP, Oracle o Salesforce lanzan un módulo REPSE nativo para México) | Baja (a corto plazo) | Muy Alto | Concentrarse en el segmento PyME que los grandes competidores nunca atenderán por su estructura de costos. Profundizar en integraciones locales (ZKTeco, biométricos económicos). |
| **Violación de datos / LFPDPPP** (Ley Federal de Protección de Datos Personales) | Baja | Muy Alto | Implementar desde el MVP: encriptación en tránsito (TLS) y en reposo (AES-256), aviso de privacidad, consentimiento explícito y política de retención de datos. Auditoría de seguridad anual. |

---

### Presupuesto Anual para Gestión de Riesgo Legal

| Concepto | Costo Anual Estimado (MXN) |
|---|---|
| Asesor EHS / Legal laboral (actualización normativa) | $72,000 |
| Auditoría de seguridad informática (pentest básico) | $30,000 |
| Registro de marca ante el IMPI | $8,000 |
| Seguro de Responsabilidad Civil Profesional | $15,000 – $25,000 |
| **Total gestión de riesgo legal/año** | **$125,000 – $135,000 MXN** |

---

## 8. Estrategia de Contenido y Marketing Digital

El plan original es 100% outbound. Esta sección agrega el canal inbound que tiene el **mayor ROI a largo plazo y menor costo marginal**.

### Por Qué el Contenido es una Ventaja Competitiva en Este Nicho

- Los gerentes EHS buscan activamente en Google: *"multa STPS contratista sin REPSE"*, *"qué documentos pedir a contratistas NOM-031"*, *"formato DC-3 vencido consecuencias"*.
- Ningún competidor tiene contenido educativo de calidad en español sobre regulación industrial mexicana.
- El contenido genera leads que ya están en modo de búsqueda (alta intención de compra).

---

### Plan de Contenido — Primeros 6 Meses

| Mes | Pilar de Contenido | Formato | Canal |
|---|---|---|---|
| 1 | "Las 5 multas más comunes de la STPS en 2025 y cómo evitarlas" | Artículo SEO + PDF descargable | Blog + LinkedIn |
| 2 | "Guía completa del REPSE para plantas industriales" | Guía larga (lead magnet) | Landing page con captura de email |
| 3 | "Video: Cómo debería verse un check-in de contratistas en caseta en 5 segundos" | Video demo + LinkedIn post | LinkedIn + YouTube |
| 4 | "Calculadora de multas STPS gratuita" | Herramienta web interactiva | Landing page + compartible |
| 5 | "Caso de éxito: Planta piloto redujo tiempo de ingreso un 70% en 14 días" | Caso de estudio (con foto de caseta real) | Blog + LinkedIn |
| 6 | "Webinar gratuito: Cómo auditar a tus contratistas sin contratar más personal" | Webinar live (60 min) | LinkedIn Events + Email list |

---

### KPIs de Contenido para los Primeros 6 Meses

| Métrica | Meta Mes 3 | Meta Mes 6 |
|---|---|---|
| Visitas orgánicas al blog/mes | 200 | 1,000 |
| Leads captados (emails) por contenido | 30 | 150 |
| Demos agendadas por inbound | 2/mes | 8/mes |
| Seguidores en LinkedIn (empresa) | 100 | 500 |
| Tasa de conversión Lead → Demo | — | 5% |

---

### Presupuesto de Marketing Digital

| Concepto | Costo Mensual (MXN) |
|---|---|
| Redacción de contenido (1–2 artículos/mes, puede ser el fundador) | $0 – $4,000 |
| LinkedIn Ads (campañas segmentadas por puesto/industria/zona) | $3,000 – $5,000 |
| Herramienta de email marketing (Brevo / Mailchimp) | $500 |
| Diseño de materiales (Canva Pro) | $300 |
| **Total marketing digital/mes** | **$3,800 – $9,800 MXN** |

---

## 9. Resumen Ejecutivo Financiero

### KPIs Objetivo — Fin de Año 1 (Escenario Base)

| Indicador | Meta Año 1 |
|---|---|
| **Clientes de pago activos** | 58 |
| **MRR** | $263,320 MXN (~$13,166 USD) |
| **ARR** | $3,159,840 MXN (~$157,992 USD) |
| **ARPA** | $4,540 MXN/mes |
| **Churn mensual** | < 2% |
| **LTV promedio** | $252,000 MXN |
| **CAC blended** | $3,800 MXN |
| **LTV/CAC** | 66× |
| **Margen bruto (real, incluyendo todos los costos)** | 43% – 50% |
| **Capital total invertido** | ~$500,000 MXN |
| **Break-even acumulado** | Mes 12 |

---

### Cronograma de Hitos Críticos

```
Mes 1–2   ████░░░░░░  Desarrollo MVP (4 módulos core). 2 Pilotos gratuitos.
Mes 3     █████░░░░░  Lanzamiento comercial. Primeros 3 clientes de pago.
Mes 4–5   ██████░░░░  5 aliados EHS activos. 10–15 clientes de pago.
Mes 6     ███████░░░  Break-even mensual. 18 clientes.
Mes 7–9   ████████░░  Desarrollo Fase 2 (PTAR, LMS, incidentes). 37 clientes.
Mes 10–12 █████████░  Break-even acumulado. 58 clientes. ARR: $3.1M MXN.
Mes 13–18 ██████████  Módulos IA (OCR, Visión EPP). 100+ clientes. ARR: $5.9M MXN.
```

---

### Tabla de Decisión: ¿Bootstrapped o Inversión Ángel?

| Criterio | Bootstrapped | Inversión Ángel ($500K–$1M MXN) |
|---|---|---|
| **Control accionario** | ✅ 100% tuyo | ❌ Cedes 15–25% |
| **Velocidad de crecimiento** | 🟡 Moderada | 🟢 Acelerada |
| **Riesgo personal** | 🔴 Alto (capital propio) | 🟡 Moderado |
| **Break-even** | Mes 12 | Mes 6–8 (con más recursos de ventas) |
| **Recomendación** | Si tienes los $500K MXN disponibles y tolerancia al riesgo | Si quieres escalar a 100 clientes antes del mes 12 |

> [!TIP]
> La recomendación es **empezar bootstrapped** con un co-fundador técnico con equity. Validar con los primeros 25 clientes de pago. Entonces buscar una ronda seed de $1–2M MXN para acelerar el canal comercial y contratar un vendedor dedicado.

---

*Documento complementario al plan de negocios principal: [SAS Seguridad e Higiene asesamo.md](./SAS%20Seguridad%20e%20Higiene%20asesamo.md)*  
*Última actualización: Octubre 2026*
