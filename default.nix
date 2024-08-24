with import <nixpkgs> {};

stdenv.mkDerivation {
  name = "reduction-go";
  buildInputs = with pkgs; [
    go
    gnumake
  ];
  hardeningDisable = [ "fortify" ];
}
