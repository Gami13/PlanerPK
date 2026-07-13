{
    description = "Automated Google Calendar management for Computer Science students at CUT WIiM.";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
        flake-parts.url = "github:hercules-ci/flake-parts";
    };

    outputs = inputs@{ flake-parts, ... }:
        flake-parts.lib.mkFlake { inherit inputs; } {
            systems = [
                "x86_64-linux"
                "aarch64-linux"
                "aarch64-darwin"
                "x86_64-darwin"
            ];

            imports = [ ];

            perSystem = { self', pkgs, ... }: {
                packages = {
                    frontend = pkgs.stdenv.mkDerivation (finalAttrs: {
                        pname = "planer-pk-frontend";
                        version = "0.1.0";
                        src = ./frontend;

                        nativeBuildInputs = [
                            pkgs.nodejs_22
                            pkgs.pnpm_9
                            pkgs.pnpmConfigHook
                        ];

                        pnpmDeps = pkgs.fetchPnpmDeps {
                            inherit (finalAttrs) pname version src;
                            hash = "sha256-8NpxHH/R6dK/sRQSpFrdAp7ASOzfUpe2jAm6r4um0ng=";
                            pnpm = pkgs.pnpm_9;
                            fetcherVersion = 3;
                        };

                        buildPhase = ''
                            runHook preBuild
                            pnpm build
                            runHook postBuild
                        '';

                        installPhase = ''
                            runHook preInstall
                            mkdir -p $out/static
                            cp -R ../static/* $out/static/
                            runHook postInstall
                        '';
                    });

                    backend = pkgs.buildGoModule {
                        pname = "planer-pk-backend";
                        version = "0.1.0";
                        src = ./backend;

                        vendorHash = "sha256-pJZWIDfjcyOsY7y4CoU46SjH0njJVoZsNlz/EluGHbk=";

                        env.CGO_ENABLED = 0;
                        ldflags = [ "-s" "-w" ];

                        subPackages = [ "." ];

                        postInstall = ''
                            mv $out/bin/PlanerPK $out/bin/planer-pk-backend
                        '';
                    };

                    default = pkgs.writeShellApplication {
                        name = "planer-pk";

                        runtimeInputs = [ pkgs.libreoffice ];

                        text = ''
                            export SERVER_PORT="''${SERVER_PORT:-8080}"
                            export API_BASE_URL="''${API_BASE_URL:-http://localhost:8080}"
                            export DATABASE_DRIVER="''${DATABASE_DRIVER:-sqlite}"

                            # Prepare the backend
                            APP_DIR="''${XDG_DATA_HOME:-$HOME/.local/share}/planer-pk"
                            mkdir -p "$APP_DIR" "$APP_DIR/data"

                            export TIMETABLE_DOWNLOAD_DESTINATION="''${TIMETABLE_DOWNLOAD_DESTINATION:-$APP_DIR/data/downloads/}"
                            export TIMETABLE_CURRENT_DESTINATION="''${TIMETABLE_CURRENT_DESTINATION:-$APP_DIR/data/current_timetable/}"
                            export JSON_INFO_LOCATION="''${JSON_INFO_LOCATION:-$APP_DIR/data/}"
                            export DATABASE_SOURCE="''${DATABASE_SOURCE:-file:$APP_DIR/data/database.db?cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite}"

                            mkdir -p "''${TIMETABLE_DOWNLOAD_DESTINATION}" "''${TIMETABLE_CURRENT_DESTINATION}"

                            CONFIG_DIR="''${XDG_CONFIG_HOME:-$HOME/.config}/planer-pk"
                            mkdir -p "$CONFIG_DIR"
                            export CALENDAR_CREDENTIALS_PATH="''${CALENDAR_CREDENTIALS_PATH:-$CONFIG_DIR/credentials.json}"

                            STATE_DIR="''${XDG_STATE_HOME:-$HOME/.local/state}/planer-pk"
                            mkdir -p "$STATE_DIR/logs"
                            export LOGS_PATH="''${LOGS_PATH:-$STATE_DIR/logs/}"

                            # Prepare the frontend
                            CACHE_DIR="''${XDG_CACHE_HOME:-$HOME/.cache}/planer-pk"
                            STATIC_DIR="$CACHE_DIR/static"

                            mkdir -p "$STATIC_DIR"
                            cp -rf ${self'.packages.frontend}/static/* "$STATIC_DIR/"
                            chmod -R u+w "$STATIC_DIR"
                            export STATIC_FILES_LOCATION="''${STATIC_FILES_LOCATION:-$STATIC_DIR}"

                            for js_file in "$STATIC_DIR"/assets/*.js; do
                                if [ -f "$js_file" ]; then
                                    # shellcheck disable=SC2016
                                    ${pkgs.gettext}/bin/envsubst '$API_BASE_URL' < "$js_file" > "$js_file.tmp"
                                    mv "$js_file.tmp" "$js_file"
                                fi
                            done

                            # Prevents LibreOffice from trying to create `/run/user/<uid>`
                            export DBUS_SESSION_BUS_ADDRESS="''${DBUS_SESSION_BUS_ADDRESS:-/dev/null}"

                            exec ${self'.packages.backend}/bin/planer-pk-backend "$@"
                        '';
                    };
                };

                apps.default = {
                    type = "app";
                    program = "${self'.packages.default}/bin/planer-pk";
                    meta.description = "Run the Planer PK backend and frontend locally";
                };

                devShells.default = pkgs.mkShell {
                    name = "planer-pk";

                    packages = with pkgs; [
                        pkg-config

                        docker
                        docker-compose

                        go
                        buf
                        sqlc
                        pnpm
                        libreoffice
                    ];
                };

                checks = {
                    inherit (self'.packages) frontend backend default;
                };
            };

            flake = {
                nixosModules.default = { config, lib, pkgs, ... }:
                let
                    cfg = config.services.planer-pk;
                    appPkgs = inputs.self.packages.${pkgs.stdenv.hostPlatform.system};
                in
                {
                    options.services.planer-pk = {
                        enable = lib.mkEnableOption "Planer PK Service";

                        port = lib.mkOption {
                            type = lib.types.port;
                            default = 8080;
                            description = "Port to listen on";
                        };

                        apiBaseUrl = lib.mkOption {
                            type = lib.types.str;
                            default = "http://localhost:8080/api";
                            description = "Base URL injected into the frontend";
                        };

                        environmentFile = lib.mkOption {
                            # Specifically a `lib.types.str` as path, because we don't want to copy sensitive data into the nix store
                            type = lib.types.nullOr lib.types.str;
                            default = null;
                            description = "Path to an .env file";
                        };

                        calendarCredentialsPath = lib.mkOption {
                            # Specifically a `lib.types.str` as path, because we don't want to copy sensitive data into the nix store
                            type = lib.types.str;
                            description = "Path to the credentials.json file";
                        };
                    };

                    config = lib.mkIf cfg.enable {
                        systemd.services.planer-pk = {
                            description = "Planer PK Backend & Frontend";
                            wantedBy = [ "multi-user.target" ];
                            after = [ "network.target" ];

                            path = [ pkgs.libreoffice ];

                            environment = {
                                SERVER_PORT = toString cfg.port;
                                API_BASE_URL = cfg.apiBaseUrl;
                                TIMETABLE_DOWNLOAD_DESTINATION = "data/downloads/";
                                TIMETABLE_CURRENT_DESTINATION = "data/current_timetable/";
                                JSON_INFO_LOCATION = "data/";
                                STATIC_FILES_LOCATION="static/";
                                CALENDAR_CREDENTIALS_PATH = "%d/credentials.json";
                                LOGS_PATH = "%L/planer-pk/";
                                DATABASE_SOURCE = "file:data/database.db?cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite";
                                DATABASE_DRIVER = "sqlite";
                                TZ = config.time.timeZone;

                                HOME = "%t/planer-pk-runtime";

                                # Prevents LibreOffice from trying to create `/run/user/<uid>`
                                DBUS_SESSION_BUS_ADDRESS = "/dev/null";
                            };

                            serviceConfig = {
                                StateDirectory = "planer-pk/data planer-pk/static";
                                WorkingDirectory = "%S/planer-pk";
                                LogsDirectory = "planer-pk";
                                RuntimeDirectory = "planer-pk-runtime";
                                ExecStart = "${appPkgs.backend}/bin/planer-pk-backend";
                                Restart = "always";
                                DynamicUser = true;
                                LoadCredential = "credentials.json:${toString cfg.calendarCredentialsPath}";
                            } // lib.optionalAttrs (cfg.environmentFile != null) {
                                EnvironmentFile = toString cfg.environmentFile;
                            };

                            preStart = ''
                                mkdir -p data/downloads data/current_timetable

                                cp -rf ${appPkgs.frontend}/static/* static/
                                chmod -R u+w static

                                for js_file in static/assets/*.js; do
                                    if [ -f "$js_file" ]; then
                                        ${pkgs.gettext}/bin/envsubst '$API_BASE_URL' < "$js_file" > "''${js_file}.tmp"
                                        mv "''${js_file}.tmp" "$js_file"
                                    fi
                                done
                            '';
                        };
                    };
                };
            };
        };
}
