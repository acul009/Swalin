{
  description = "A very basic flake";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {

      devShells.x86_64-linux.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go_1_21
          gotools
          glibc
          pkg-config
          gcc
          libGL.dev
          xorg.libX11.dev
          xorg.libXcursor
          xorg.libXrandr
          xorg.libXinerama
          xorg.libXi.dev
          xorg.libXxf86vm
          glfw
        ];



        shellHook = "zsh;exit;";
      };
    };
}
