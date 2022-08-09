{
  description = "Terraform Provider for Talos";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { flake-utils, nixpkgs, self }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go_1_18
            pkgs.go-outline
            pkgs.go-tools
            pkgs.gopls
            pkgs.goreleaser
            pkgs.kubectl
            pkgs.talosctl
            pkgs.terraform
          ];
        };
        formatter = pkgs.nixpkgs-fmt;
      }
    );
}
