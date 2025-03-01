{
  description = "mail";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachSystem 
    [ "aarch64-darwin" ] 
    (system: 
      let 
        pkgs = nixpkgs.legacyPackages.${system};
      in
    {
      devShells.default = with pkgs; mkShell {
          name = "mail";
          buildInputs = [
            go
            gopls
          ];
      };
    });
}
