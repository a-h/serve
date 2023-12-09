{
  description = "serve";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/23.11";
  };

  outputs = { self, nixpkgs }:
    let
      # Systems supported
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      # Helper to provide system-specific attributes
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        system = system;
        pkgs = import nixpkgs {
          inherit system;
        };
      });

      # Build for the app.
      serve = pkgs: pkgs.buildGo121Module {
        name = "serve";
        src = ./.;
        vendorHash = null; # Use vendored deps.
      };

    in
    {
      # `nix build` builds the app.
      packages = forAllSystems ({ system, pkgs }: {
        default = serve pkgs;
      });
      # `nix develop` provides a shell containing required tools.
      devShell = forAllSystems
        ({ system, pkgs }:
          pkgs.mkShell {
            buildInputs = [
              pkgs.go_1_21
            ];
          });
    };
}
