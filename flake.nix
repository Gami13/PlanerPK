{
    description = "Server for automated fetching, parsing, and embedding CUT timetables into calendars or other formats.";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

        gomod2nix = {
            url = "github:nix-community/gomod2nix";
            inputs.nixpkgs.follows = "nixpkgs";
            inputs.flake-utils.follows = "flake-utils";
        };

        flake-utils = {
            url = "github:numtide/flake-utils";
        };
    };

    outputs = { self, nixpkgs, flake-utils, gomod2nix, ... }:
    flake-utils.lib.eachDefaultSystem (system:
    let
        pkgs = nixpkgs.legacyPackages.${system};
        goMod = gomod2nix.legacyPackages.${system};
        projectName = "google-calendar-timetable";
    in
    {
        devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
                libreoffice
            ];

            nativeBuildInputs = with pkgs; [
                go
                goMod.gomod2nix
                (goMod.mkGoEnv { pwd = ./.; })
            ];

            shellHook = ''
                printf '\x1b[36m\x1b[1m\x1b[4mTime to develop ${projectName}!\x1b[0m\n\n'
            '';
        };

        packages.default = goMod.buildGoApplication {
            pname = projectName;
            version = "0.1.0";
            src = ./.;
            modules = ./gomod2nix.toml;
        };
    });
}
