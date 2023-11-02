{
  description = "Be a lazygit{hub,lab}";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";

  # Flake outputs
  outputs = { self, nixpkgs, gomod2nix }:
    let
      # Systems supported
      allSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      # Helper to provide system-specific attributes
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
      });
    in
      {
        # Development environment output
        devShells = forAllSystems ({ pkgs }: {
          default = pkgs.mkShell {
            # The Nix packages provided in the environment
            packages = with pkgs; [
              gnumake
              go_1_20
              godef # jump to definition in editors
              golangci-lint # fast linter runners
              gotools # Go tools like goimports, godoc, and others
              gomod2nix.packages.${system}.default
              gopls # go language server for using lsp plugins
            ];
          };
        });
      };
}
