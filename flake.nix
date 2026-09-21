{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forEachSystem = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      devShells = forEachSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              pkgs.go_1_27
              pkgs.gopls
              pkgs.swaks
              pkgs.inetutils
              pkgs.openssl
            ];

            shellHook = ''
              echo "devShell is active"
              echo "-----"
              echo "Golang version: $(go version)"
            '';
          };
        }
      );
    };
}
