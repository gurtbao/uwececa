{
  description = "Personal Blogs For UW ECE Students.";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
  };

  outputs = inputs @ {self, ...}: let
    supportedSystems = ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];
    forAllSystems = inputs.nixpkgs.lib.genAttrs supportedSystems;
    nixpkgsFor = forAllSystems (system: inputs.nixpkgs.legacyPackages.${system});
  in {
    devShells = forAllSystems (system: let
      pkgs = nixpkgsFor.${system};
    in {
      default = pkgs.mkShell {
        nativeBuildInputs = [
          pkgs.go
          pkgs.go-tools
          pkgs.gotools
          pkgs.gopls
          pkgs.nixos-shell
          pkgs.coreutils
          pkgs.gcc
        ];

        env.CGO_ENABLED = 1;
      };
    });
  };
}
