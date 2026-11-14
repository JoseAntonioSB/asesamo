#!/usr/bin/env bash
# ==============================================================================
# ASESAMO — Industrial Safety, Access & EHS Management Platform
# File: start.sh
# Description: Unified startup, orchestration, health-check and management script.
#              Supports Docker Compose (default) and Local Native execution.
# ==============================================================================

set -o pipefail

# Project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment configuration if present
if [ -f "$SCRIPT_DIR/.env" ]; then
    # shellcheck disable=SC1090
    set -a
    source "$SCRIPT_DIR/.env"
    set +a
fi

WEB_PORT="${WEB_PORT:-3100}"
API_PORT="${API_PORT:-8080}"

# PID and log files for local execution
PID_DIR="$SCRIPT_DIR/.pids"
LOG_DIR="$SCRIPT_DIR/.logs"
API_PID_FILE="$PID_DIR/api.pid"
WEB_PID_FILE="$PID_DIR/web.pid"
API_LOG_FILE="$LOG_DIR/api.log"
WEB_LOG_FILE="$LOG_DIR/web.log"

# Colors & Formatting
BOLD="\033[1m"
DIM="\033[2m"
CYAN="\033[36m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
MAGENTA="\033[35m"
BLUE="\033[34m"
RESET="\033[0m"

print_banner() {
    cat << "EOF"
    ___    _____ ______ _____ ___     __  _______ 
   /   |  / ___// ____// ___//   |   /  |/  / __ \
  / /| |  \__ \/ __/   \__ \/ /| |  / /|_/ / / / /
 / ___ | ___/ / /___  ___/ / ___ | / /  / / /_/ / 
/_/  |_|/____/_____/ /____/_/  |_|/_/  /_/\____/  
EOF
    echo -e "${DIM}================================================================${RESET}"
    echo -e "${CYAN}${BOLD} ASESAMO Industrial Platform — Control & Startup Engine${RESET}"
    echo -e "${DIM} Security, Access Control, Work Permits, EHS, DC-3 & Health${RESET}"
    echo -e "${DIM}================================================================${RESET}\n"
}

log_info() {
    echo -e "${BLUE}${BOLD}[INFO]${RESET} $1"
}

log_success() {
    echo -e "${GREEN}${BOLD}[OK]${RESET} $1"
}

log_warn() {
    echo -e "${YELLOW}${BOLD}[WARN]${RESET} $1"
}

log_error() {
    echo -e "${RED}${BOLD}[ERROR]${RESET} $1"
}

show_help() {
    print_banner
    echo -e "${BOLD}USO:${RESET}"
    echo -e "  ./start.sh [OPCIÓN]\n"
    echo -e "${BOLD}OPCIONES PRINCIPALES:${RESET}"
    echo -e "  ${GREEN}(sin argumentos)${RESET}   Levanta todo el stack con Docker Compose (Base de datos, migraciones, seeds, API y Web)"
    echo -e "  ${GREEN}-d, --docker${RESET}        Modo Docker Compose explícito (build + up de todos los servicios)"
    echo -e "  ${GREEN}-l, --local${RESET}         Modo Local / Desarrollo nativo (Postgres en Docker/host + Go API + Next.js)"
    echo -e "  ${GREEN}-s, --stop${RESET}          Detiene todos los contenedores y procesos locales en ejecución"
    echo -e "  ${GREEN}--status${RESET}            Muestra el estado de salud de los servicios, puertos y PIDs"
    echo -e "  ${GREEN}--seed${RESET}              Ejecuta todos los datasets de prueba (Fases 1, 3 y 4)"
    echo -e "  ${GREEN}--migrate${RESET}           Aplica todas las migraciones SQL pendientes"
    echo -e "  ${GREEN}--logs${RESET}              Muestra los logs en vivo de los servicios"
    echo -e "  ${GREEN}-h, --help${RESET}          Muestra esta ayuda interactiva\n"
    echo -e "${BOLD}EJEMPLOS:${RESET}"
    echo -e "  ./start.sh               # Levanta el sistema completo con Docker"
    echo -e "  ./start.sh --local       # Levanta el sistema en modo desarrollo nativo"
    echo -e "  ./start.sh --status      # Verifica el estado de puertos y servicios"
    echo -e "  ./start.sh --stop        # Detiene todo de manera limpia\n"
}

check_command() {
    command -v "$1" >/dev/null 2>&1
}

# Determine Docker Compose command variant
get_compose_cmd() {
    if docker compose version >/dev/null 2>&1; then
        echo "docker compose"
    elif check_command docker-compose; then
        echo "docker-compose"
    else
        echo ""
    fi
}

wait_for_http() {
    local url="$1"
    local name="$2"
    local max_retries="${3:-30}"
    local delay="${4:-2}"
    local count=0

    echo -ne "  Esperando a que ${BOLD}$name${RESET} responda en ${CYAN}$url${RESET} "
    while [ $count -lt $max_retries ]; do
        if check_command curl; then
            if curl -s -f -o /dev/null -m 2 "$url"; then
                echo -e " ${GREEN}✔ Listo${RESET}"
                return 0
            fi
        elif check_command wget; then
            if wget -q --spider --timeout=2 "$url"; then
                echo -e " ${GREEN}✔ Listo${RESET}"
                return 0
            fi
        fi
        echo -ne "${DIM}.${RESET}"
        sleep $delay
        count=$((count + 1))
    done

    echo -e " ${RED}✖ Tiempo de espera agotado${RESET}"
    return 1
}

display_credentials() {
    echo -e "\n${BOLD}${GREEN}================================================================${RESET}"
    echo -e "${BOLD}${GREEN} ✔ ASESAMO ESTÁ LISTO Y OPERATIVO${RESET}"
    echo -e "${BOLD}${GREEN}================================================================${RESET}\n"

    echo -e "${BOLD}Puntos de Acceso:${RESET}"
    echo -e "  • ${BOLD}Frontend Web:${RESET}       ${CYAN}http://localhost:${WEB_PORT}${RESET}"
    echo -e "  • ${BOLD}Backend REST API:${RESET}   ${CYAN}http://localhost:${API_PORT}/api/v1${RESET}"
    echo -e "  • ${BOLD}Healthcheck API:${RESET}    ${CYAN}http://localhost:${API_PORT}/api/v1/health${RESET}"
    echo -e "  • ${BOLD}SSE Real-Time Hub:${RESET}  ${CYAN}http://localhost:${API_PORT}/api/v1/realtime/events${RESET}"
    echo -e "  • ${BOLD}Base de Datos:${RESET}      ${CYAN}localhost:5432 (asesamo_db / asesamo_admin)${RESET}\n"

    echo -e "${BOLD}Cuentas Semilla Preconfiguradas (Password para todas: ${YELLOW}Password123!${RESET}):${RESET}"
    echo -e "  ┌──────────────────┬──────────────────────────────────────┬─────────────────────────┐"
    echo -e "  │ ${BOLD}Rol${RESET}              │ ${BOLD}Email${RESET}                                │ ${BOLD}Usuario${RESET}                 │"
    echo -e "  ├──────────────────┼──────────────────────────────────────┼─────────────────────────┤"
    echo -e "  │ Administrador    │ ${CYAN}admin@asesamo.com${RESET}                    │ Administrador de Planta │"
    echo -e "  │ Seguridad / EHS  │ ${CYAN}ehs@asesamo.com${RESET}                      │ Ing. Laura Sanchez      │"
    echo -e "  │ Vigilancia Caseta│ ${CYAN}guard@asesamo.com${RESET}                    │ Oficial Roberto Caseta  │"
    echo -e "  │ Contratista Admin│ ${CYAN}contratista@mantenimientonorte.com${RESET}   │ Ing. Fernando Garza     │"
    echo -e "  │ Trabajador Op.   │ ${CYAN}carlos.mendoza@mantenimientonorte.com${RESET}│ Carlos Mendoza          │"
    echo -e "  └──────────────────┴──────────────────────────────────────┴─────────────────────────┘\n"

    echo -e "${DIM}Para ver el estado de la plataforma:${RESET} ${BOLD}./start.sh --status${RESET}"
    echo -e "${DIM}Para detener los servicios:${RESET}          ${BOLD}./start.sh --stop${RESET}\n"
}

start_docker() {
    print_banner
    log_info "Iniciando ASESAMO mediante Docker Compose..."

    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)
    if [ -z "$COMPOSE_CMD" ]; then
        log_error "Docker o Docker Compose no están disponibles en el sistema."
        log_warn "Puedes instalar Docker Desktop o intentar arrancar en modo local: ./start.sh --local"
        exit 1
    fi

    log_info "Compilando imágenes y desplegando contenedores..."
    $COMPOSE_CMD up -d --build

    log_info "Verificando disponibilidad de servicios..."
    
    # 1. Wait for DB and migrations
    echo -ne "  Verificando contenedor de base de datos..."
    local db_ready=0
    for i in $(seq 1 20); do
        if $COMPOSE_CMD exec -T db pg_isready -U asesamo_admin -d asesamo_db >/dev/null 2>&1; then
            db_ready=1
            break
        fi
        sleep 1
    done

    if [ $db_ready -eq 1 ]; then
        echo -e " ${GREEN}✔ Saludable${RESET}"
    else
        echo -e " ${YELLOW}⚠ Tomando más tiempo de lo esperado${RESET}"
    fi

    # 2. Check API health
    wait_for_http "http://localhost:${API_PORT}/api/v1/health" "API Go Clean Architecture" 25 2 || log_warn "API tardó en responder. Verifique con ./start.sh --logs"

    # 3. Check Web health
    wait_for_http "http://localhost:${WEB_PORT}/login" "Frontend Next.js 14" 30 2 || log_warn "Frontend web tardó en responder. Verifique con ./start.sh --logs"

    display_credentials
}

start_local() {
    print_banner
    log_info "Iniciando ASESAMO en Modo Desarrollo Local Nativo..."

    mkdir -p "$PID_DIR" "$LOG_DIR"

    # Check prerequisites
    if ! check_command go; then
        log_error "Go no está instalado. Se requiere Go 1.22+ para ejecutar asesamo-api."
        exit 1
    fi

    if ! check_command node || ! check_command npm; then
        log_error "Node.js / npm no están instalados. Se requiere Node.js 18+ para ejecutar asesamo-web."
        exit 1
    fi

    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)

    # 1. Database readiness check
    log_info "Verificando conexión a PostgreSQL (localhost:5432)..."
    local db_connected=0

    if check_command pg_isready && pg_isready -h localhost -p 5432 -U asesamo_admin >/dev/null 2>&1; then
        db_connected=1
        log_success "Base de datos PostgreSQL detectada en localhost:5432."
    elif [ -n "$COMPOSE_CMD" ]; then
        log_warn "Postgres local no detectado en 5432. Levantando servicio 'db' con Docker Compose..."
        $COMPOSE_CMD up -d db migrate seeder
        sleep 3
        db_connected=1
    fi

    if [ $db_connected -eq 0 ]; then
        log_error "No se pudo conectar a PostgreSQL en localhost:5432 y Docker no está disponible."
        log_error "Por favor inicia PostgreSQL manualmente con las credenciales de .env."
        exit 1
    fi

    # 2. Start Go API
    log_info "Compilando e iniciando Backend Go (asesamo-api)..."
    cd "$SCRIPT_DIR/asesamo-api"

    # Kill any stale API process
    if [ -f "$API_PID_FILE" ]; then
        local old_pid
        old_pid=$(cat "$API_PID_FILE")
        if kill -0 "$old_pid" 2>/dev/null; then
            log_warn "Deteniendo instancia previa de API (PID: $old_pid)..."
            kill "$old_pid" 2>/dev/null || true
            sleep 1
        fi
        rm -f "$API_PID_FILE"
    fi

    # Compile binary for fast startup
    mkdir -p bin
    go build -o bin/api cmd/api/main.go
    APP_ENV=development \
    APP_PORT="${API_PORT}" \
    DATABASE_URL="postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable" \
    JWT_SECRET="super_secret_jwt_key_min_32_characters_long_12345" \
    CORS_ALLOWED_ORIGINS="http://localhost:${WEB_PORT},http://127.0.0.1:${WEB_PORT},http://localhost:3000,http://127.0.0.1:3000" \
    STORAGE_PATH="$SCRIPT_DIR/uploads" \
    ./bin/api > "$API_LOG_FILE" 2>&1 &

    local api_pid=$!
    echo "$api_pid" > "$API_PID_FILE"
    log_success "asesamo-api corriendo en segundo plano (PID: $api_pid, Log: .logs/api.log)"

    # 3. Start Next.js Frontend
    log_info "Iniciando Frontend Next.js (asesamo-web)..."
    cd "$SCRIPT_DIR/asesamo-web"

    # Kill any stale Web process
    if [ -f "$WEB_PID_FILE" ]; then
        local old_web_pid
        old_web_pid=$(cat "$WEB_PID_FILE")
        if kill -0 "$old_web_pid" 2>/dev/null; then
            log_warn "Deteniendo instancia previa de Web (PID: $old_web_pid)..."
            kill "$old_web_pid" 2>/dev/null || true
            sleep 1
        fi
        rm -f "$WEB_PID_FILE"
    fi

    if [ ! -d "node_modules" ]; then
        log_info "Instalando dependencias de npm en asesamo-web..."
        npm install
    fi

    PORT="${WEB_PORT}" \
    NEXT_PUBLIC_API_URL="http://localhost:${API_PORT}/api/v1" \
    npm run dev -- -p "${WEB_PORT}" > "$WEB_LOG_FILE" 2>&1 &

    local web_pid=$!
    echo "$web_pid" > "$WEB_PID_FILE"
    log_success "asesamo-web corriendo en segundo plano (PID: $web_pid, Puerto: $WEB_PORT, Log: .logs/web.log)"

    cd "$SCRIPT_DIR"

    # Wait for HTTP endpoints
    wait_for_http "http://localhost:${API_PORT}/api/v1/health" "API Go (Local)" 20 1 || log_warn "API tardó en iniciar. Revise .logs/api.log"
    wait_for_http "http://localhost:${WEB_PORT}/login" "Next.js Web (Local)" 25 1 || log_warn "Web tardó en iniciar. Revise .logs/web.log"

    display_credentials

    # Trap clean exit on interactive runs if requested
    log_info "Para seguir los logs en tiempo real: ${BOLD}tail -f .logs/api.log .logs/web.log${RESET}"
}

stop_all() {
    print_banner
    log_info "Deteniendo todos los servicios de ASESAMO..."

    # 1. Stop local processes
    if [ -f "$API_PID_FILE" ]; then
        local api_pid
        api_pid=$(cat "$API_PID_FILE")
        if kill -0 "$api_pid" 2>/dev/null; then
            log_info "Deteniendo asesamo-api local (PID: $api_pid)..."
            kill "$api_pid" 2>/dev/null || true
        fi
        rm -f "$API_PID_FILE"
    fi

    if [ -f "$WEB_PID_FILE" ]; then
        local web_pid
        web_pid=$(cat "$WEB_PID_FILE")
        if kill -0 "$web_pid" 2>/dev/null; then
            log_info "Deteniendo asesamo-web local (PID: $web_pid)..."
            kill "$web_pid" 2>/dev/null || true
        fi
        rm -f "$WEB_PID_FILE"
    fi

    # Clean any rogue node/go listeners on standard ports
    if check_command lsof; then
        local p_api
        p_api=$(lsof -ti:"${API_PORT}" 2>/dev/null || true)
        if [ -n "$p_api" ]; then
            log_info "Liberando puerto ${API_PORT} (PID: $p_api)..."
            kill "$p_api" 2>/dev/null || true
        fi
        local p_web
        p_web=$(lsof -ti:"${WEB_PORT}" 2>/dev/null || true)
        if [ -n "$p_web" ]; then
            log_info "Liberando puerto ${WEB_PORT} (PID: $p_web)..."
            kill "$p_web" 2>/dev/null || true
        fi
    fi

    # 2. Stop docker containers
    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)
    if [ -n "$COMPOSE_CMD" ]; then
        log_info "Deteniendo contenedores Docker Compose..."
        $COMPOSE_CMD down --remove-orphans 2>/dev/null || true
    fi

    log_success "Todos los servicios han sido detenidos correctamente."
}

show_status() {
    print_banner
    log_info "Comprobando estado de la plataforma ASESAMO...\n"

    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)

    echo -e "${BOLD}1. Estado de Contenedores Docker:${RESET}"
    if [ -n "$COMPOSE_CMD" ]; then
        $COMPOSE_CMD ps
    else
        echo -e "  ${DIM}Docker Compose no está disponible.${RESET}"
    fi

    echo -e "\n${BOLD}2. Estado de Procesos Locales:${RESET}"
    if [ -f "$API_PID_FILE" ]; then
        local api_pid
        api_pid=$(cat "$API_PID_FILE")
        if kill -0 "$api_pid" 2>/dev/null; then
            echo -e "  • ${BOLD}API Go:${RESET}  ${GREEN}ACTIVO${RESET} (PID: $api_pid)"
        else
            echo -e "  • ${BOLD}API Go:${RESET}  ${RED}INACTIVO${RESET} (PID huérfano: $api_pid)"
        fi
    else
        echo -e "  • ${BOLD}API Go:${RESET}  ${DIM}No registrado en modo local${RESET}"
    fi

    if [ -f "$WEB_PID_FILE" ]; then
        local web_pid
        web_pid=$(cat "$WEB_PID_FILE")
        if kill -0 "$web_pid" 2>/dev/null; then
            echo -e "  • ${BOLD}Web Next:${RESET} ${GREEN}ACTIVO${RESET} (PID: $web_pid)"
        else
            echo -e "  • ${BOLD}Web Next:${RESET} ${RED}INACTIVO${RESET} (PID huérfano: $web_pid)"
        fi
    else
        echo -e "  • ${BOLD}Web Next:${RESET} ${DIM}No registrado en modo local${RESET}"
    fi

    echo -e "\n${BOLD}3. Puertos de Red y Healthchecks HTTP:${RESET}"
    # API Health
    if check_command curl; then
        local api_health
        api_health=$(curl -s -o /dev/null -w "%{http_code}" -m 2 "http://localhost:${API_PORT}/api/v1/health" 2>/dev/null || echo "000")
        api_health="${api_health: -3}"
        if [ "$api_health" = "200" ]; then
            echo -e "  • Puerto ${API_PORT} (API):  ${GREEN}ONLINE (HTTP 200 OK)${RESET}"
        else
            echo -e "  • Puerto ${API_PORT} (API):  ${RED}OFFLINE${RESET}"
        fi

        local web_health
        web_health=$(curl -s -o /dev/null -w "%{http_code}" -m 2 "http://localhost:${WEB_PORT}/login" 2>/dev/null || echo "000")
        web_health="${web_health: -3}"
        if [ "$web_health" = "200" ]; then
            echo -e "  • Puerto ${WEB_PORT} (Web):  ${GREEN}ONLINE (HTTP 200 OK)${RESET}"
        else
            echo -e "  • Puerto ${WEB_PORT} (Web):  ${RED}OFFLINE${RESET}"
        fi
    fi
    echo ""
}

run_seeds() {
    print_banner
    log_info "Ejecutando siembra de datos de prueba en la base de datos..."
    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)

    if [ -n "$COMPOSE_CMD" ] && $COMPOSE_CMD ps db --format '{{.State}}' 2>/dev/null | grep -q "running"; then
        for seed_file in db/seeds/*.sql; do
            log_info "Aplicando seed: $seed_file..."
            $COMPOSE_CMD exec -T -e PGPASSWORD=asesamo_secret_dev db psql -U asesamo_admin -d asesamo_db < "$seed_file"
        done
        log_success "Todos los datos de prueba han sido aplicados con éxito."
    else
        log_error "El contenedor de base de datos no está corriendo. Ejecuta './start.sh' primero."
        exit 1
    fi
}

run_migrations() {
    print_banner
    log_info "Aplicando migraciones pendientes..."
    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)

    if [ -n "$COMPOSE_CMD" ]; then
        $COMPOSE_CMD run --rm migrate up
        log_success "Migraciones completadas."
    else
        log_error "Docker Compose es necesario para ejecutar migrate:migrate."
        exit 1
    fi
}

show_logs() {
    local COMPOSE_CMD
    COMPOSE_CMD=$(get_compose_cmd)
    if [ -f "$API_LOG_FILE" ] || [ -f "$WEB_LOG_FILE" ]; then
        log_info "Mostrando logs locales..."
        tail -f "$LOG_DIR"/*.log
    elif [ -n "$COMPOSE_CMD" ]; then
        log_info "Mostrando logs de Docker Compose..."
        $COMPOSE_CMD logs -f
    else
        log_error "No se encontraron logs activos."
    fi
}

# ------------------------------------------------------------------------------
# Entrypoint Router
# ------------------------------------------------------------------------------
case "${1:-}" in
    -h|--help|help)
        show_help
        ;;
    -d|--docker|docker)
        start_docker
        ;;
    -l|--local|local)
        start_local
        ;;
    -s|--stop|stop|down)
        stop_all
        ;;
    --status|status)
        show_status
        ;;
    --seed|seed)
        run_seeds
        ;;
    --migrate|migrate)
        run_migrations
        ;;
    --logs|logs)
        show_logs
        ;;
    "")
        # Default action: Docker Compose if available, otherwise suggest local
        start_docker
        ;;
    *)
        echo -e "${RED}Opción desconocida: $1${RESET}\n"
        show_help
        exit 1
        ;;
esac
