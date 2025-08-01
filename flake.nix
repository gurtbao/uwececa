{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";

    flake-parts.url = "github:hercules-ci/flake-parts";

    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    alejandra = {
      url = "github:kamadorueda/alejandra/3.0.0";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {self, ...}:
    inputs.flake-parts.lib.mkFlake {inherit inputs;} (toplevel @ {withSystem, ...}: let
      getPackages = system:
        import inputs.nixpkgs {
          localSystem = system;
          config = {
            allowUnfree = true;
            allowAliases = true;
          };

          overlays = [
            inputs.rust-overlay.overlays.default
          ];
        };
    in {
      systems = ["aarch64-linux" "aarch64-linux" "x86_64-linux"];

      perSystem = {
        config,
        self',
        inputs',
        pkgs,
        system,
        ...
      }: {
        _module.args.pkgs = getPackages system;

        devShells.default = pkgs.mkShell rec {
          buildInputs = with pkgs; [
            rust-bin.stable.latest.default
            sqlx-cli
            postgresql
            openssl
            pkg-config
          ];
        };

        formatter = inputs'.alejandra.packages.default;
      };
    });
}
