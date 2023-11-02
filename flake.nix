{
  description = "Be a lazygit{hub,lab}";

  # Flake inputs
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          # The current default sdk for macOS fails to compile go projects, so
          # we use a newer one for now. This has no effect on other platforms.
          callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
        in
          {
            packages.default = callPackage ./. {
              inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
            };
            devShells.default = callPackage ./shell.nix {
              inherit (gomod2nix.legacyPackages.${system}) buildGoApplication mkGoEnv gomod2nix;
              packages = with pkgs; [
                gnumake
                go_1_20
                godef # jump to definition in editors
                golangci-lint # fast linter runners
                gotools # Go tools like goimports, godoc, and others
                gopls # go language server for using lsp plugins
              ];
            };
          })
    );
}
