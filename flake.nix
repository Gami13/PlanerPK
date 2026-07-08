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

                    # TODO: A working convenience wrapper to run them both locally without NixOS
                    default = pkgs.symlinkJoin {
                        name = "planer-pk-combined";
                        paths = [ self'.packages.backend self'.packages.frontend ];
                    };
                };

                devShells.default = pkgs.mkShell {
                    name = "planer-pk";

                    packages = with pkgs; [
                        pkg-config

                        docker
                        docker-compose

                        buf
                        sqlc
                        pnpm
                        libreoffice
                    ];
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
                            type = lib.types.nullOr lib.types.path;
                            default = null;
                            description = "Path to an .env file";
                        };

                        calendarCredentialsPath = lib.mkOption {
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
                                CALENDAR_CREDENTIALS_PATH = "%d/credentials.json";
                                TIMETABLE_DOWNLOAD_DESTINATION = "./downloads/";
                                TIMETABLE_CURRENT_DESTINATION = "./current_timetable/";
                                DATABASE_SOURCE = "file:database.db?cache:shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite";
                                DATABASE_DRIVER = "sqlite";
                                TZ = config.time.timeZone;
                            };

                            serviceConfig = {
                                StateDirectory = "planer-pk/app planer-pk/data";
                                WorkingDirectory = "%S/planer-pk/app";
                                ExecStart = "${appPkgs.backend}/bin/planer-pk-backend";
                                Restart = "always";
                                DynamicUser = true;
                                LoadCredential = "credentials.json:${toString cfg.calendarCredentialsPath}";
                            } // lib.optionalAttrs (cfg.environmentFile != null) {
                                EnvironmentFile = toString cfg.environmentFile;
                            };

                            preStart = ''
                                mkdir -p ../data/current_timetable ../data/downloads ../data/logs

                                ln -sf ../data/current_timetable ./current_timetable
                                ln -sf ../data/downloads ./downloads
                                ln -sf ../data/logs ./logs
                                ln -sf ../data/database.db ./database.db

                                mkdir -p static
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
