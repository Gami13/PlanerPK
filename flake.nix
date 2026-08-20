{
  description = "Automated Google Calendar management for Computer Science students at CUT WIiM.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ flake-parts, ... }:
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
              pkgs.nodejs
              pkgs.pnpm
              pkgs.pnpmConfigHook
            ];

            pnpmDeps = pkgs.fetchPnpmDeps {
              inherit (finalAttrs) pname version src;
              hash = "sha256-TolB/R3ox3bnz4R4i22ZWzi75xy7QzSDR/XJXXJ5+gU=";
              pnpm = pkgs.pnpm;
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

            vendorHash = "sha256-QVjO2CY1tB1J0x9oQy5+6xBnX59XIlEoXsU5VUCVbi0=";

            env.CGO_ENABLED = 0;
            ldflags = [
              "-s"
              "-w"
            ];

            subPackages = [ "." ];

            postInstall = ''
              mv $out/bin/PlanerPK $out/bin/planer-pk-backend
            '';
          };

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

        formatter = pkgs.writeShellApplication {
          name = "planer-pk-formatter";

          runtimeInputs = with pkgs; [
            gofumpt
            nixfmt-rfc-style
          ];

          text = ''
            find . -type f -name '*.nix' -exec ${pkgs.nixfmt}/bin/nixfmt {} +
            ${pkgs.gofumpt}/bin/gofumpt -l -w ./backend
          '';
        };

        checks = {
          inherit (self'.packages) frontend backend;

          # Is everything formatted
          formatting =
            pkgs.runCommand "check-formatting"
              {
                nativeBuildInputs = with pkgs; [
                  gofumpt
                  nixfmt-rfc-style
                ];
              }
              ''
                # Copy the source code into the build sandbox
                cp -r ${./.} src
                cd src

                echo "Checking Nix formatting..."
                find . -type f -name '*.nix' -exec ${pkgs.nixfmt}/bin/nixfmt --check {} +

                echo "Checking Go formatting..."
                UNFORMATTED=$(${pkgs.gofumpt}/bin/gofumpt -l ./backend)

                if [ -n "$UNFORMATTED" ]; then
                    echo "Error: The following Go files need to be formatted:"
                    echo "$UNFORMATTED"
                    echo "Run 'nix fmt' to fix this."
                    exit 1
                fi

                touch $out
              '';
        };
      };

      flake = {
        nixosModules.default =
          {
            config,
            lib,
            pkgs,
            ...
          }:
          let
            cfg = config.services.planer-pk;
            appPkgs = inputs.self.packages.${pkgs.stdenv.hostPlatform.system};
          in
          {
            options.services.planer-pk = {
              enable = lib.mkEnableOption "Planer PK Service";

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
                  API_BASE_URL = cfg.apiBaseUrl;

                  DATA_DIR = "%S/planer-pk/data";
                  TZ = config.time.timeZone;

                  HOME = "%t/planer-pk-runtime";

                  # Prevents LibreOffice from trying to create `/run/user/<uid>`
                  DBUS_SESSION_BUS_ADDRESS = "/dev/null";
                };

                serviceConfig = {
                  StateDirectory = "planer-pk/data";
                  WorkingDirectory = "%S/planer-pk";
                  RuntimeDirectory = "planer-pk-runtime";
                  ExecStart = "${appPkgs.backend}/bin/planer-pk-backend";
                  Restart = "always";
                  DynamicUser = true;
                  LoadCredential = "credentials.json:${toString cfg.calendarCredentialsPath}";
                }
                // lib.optionalAttrs (cfg.environmentFile != null) {
                  EnvironmentFile = toString cfg.environmentFile;
                };

                preStart = ''
                  mkdir -p "$DATA_DIR" static

                  cp -rf ${appPkgs.frontend}/static/* static/
                  chmod -R u+w static

                  for js_file in static/assets/*.js; do
                      if [ -f "$js_file" ]; then
                          ${pkgs.gettext}/bin/envsubst '$API_BASE_URL' < "$js_file" > "''${js_file}.tmp"
                          mv "''${js_file}.tmp" "$js_file"
                      fi
                  done

                  cp "$CREDENTIALS_DIRECTORY/credentials.json" "$DATA_DIR/credentials.json"
                  chmod 600 "$DATA_DIR/credentials.json"
                '';
              };
            };
          };
      };
    };
}
