{
  description = "serve";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/25.11";
  };

  outputs = { self, nixpkgs }:
    let
      # Systems supported.
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      # Helper to provide system-specific attributes.
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        system = system;
        pkgs = import nixpkgs {
          inherit system;
        };
      });

      # Build for the app.
      serve = pkgs: pkgs.buildGoModule {
        name = "serve";
        src = ./.;
        vendorHash = null; # Use vendored deps.
      };

      # Container image for the app.
      image = pkgs: pkgs.dockerTools.buildLayeredImage {
        name = "ghcr.io/a-h/serve";
        tag = "latest";
        contents = [
          (serve pkgs)
        ];
        config = {
          Cmd = [ "/bin/serve" ];
          Env = [
            "SERVE_ADDR=:8080"
            "SERVE_DIR=/data"
            "SERVE_READ_ONLY=true"
          ];
          Expose = [ "8080/tcp" ];
        };
      };
    in
    {
      # `nix build` builds the app.
      packages = forAllSystems ({ system, pkgs }: {
        default = serve pkgs;
        serve = serve pkgs;
        image = image pkgs;
      });
      # `nix develop` provides a shell containing required tools.
      devShell = forAllSystems ({ system, pkgs }:
        pkgs.mkShell {
          buildInputs = [
            pkgs.curl
            pkgs.go
            pkgs.skopeo
          ];
        });
    };
}
