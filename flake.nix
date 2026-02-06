{
  description = "Nix Flake for Go Development for Gotify Plugins";
  
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };
  outputs = { self, nixpkgs }: 
  let 
    pkgs = import nixpkgs { system = "x86_64-linux"; config.allowUnfree = true; };
  in
  {
    devShells."x86_64-linux".default = pkgs.mkShell {
      name = "go-dev-env";
      
      # Dev enviroment dependancies
      packages = [
        # VS Code and Extensions
        (pkgs.vscode-with-extensions.override { 
          vscodeExtensions = with pkgs.vscode-extensions; [
            bbenoist.nix
            golang.go
          ] ++ pkgs.vscode-utils.extensionsFromVscodeMarketplace [
          ];
        })
        pkgs.go
        pkgs.wget
        pkgs.gnumake
      ];

      # Add Dependancies of other packages
      inputsFrom = [];

      shellHook = ''
        code .;
        exit;
      '';

      # Enviroment Variables
      # test = "AAAAAAA";
    };

  };
}
