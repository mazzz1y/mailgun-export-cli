{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs systems (
          system:
          let
            pkgs = import nixpkgs { inherit system; };
          in
          f pkgs
        );
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "mailgun-export-csv";
          version = "1.0.0";
          src = ./.;
          vendorHash = "sha256-ucco8IL6RBEVG/9/zg2Ox60fy7bLljn+ncdBhWLT7Yk=";
          meta = {
            description = "Export email events from Mailgun to CSV";
            mainProgram = "mailgun-export-csv";
          };
        };
      });
    };
}
