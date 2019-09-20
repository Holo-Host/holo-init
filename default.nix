{ pkgs ? import ./pkgs.nix {} }:

with pkgs;

stdenv.mkDerivation {
  name = "holo-init";

  nativeBuildInputs = [ makeWrapper ];
  buildInputs = [ python3 ];

  buildCommand = ''
    makeWrapper ${python3}/bin/python3 $out/bin/holo-init \
      --add-flags ${./holo-init.py} \
      --prefix PATH : ${lib.makeBinPath [ holochain-cli zerotierone ]}
  '';
}
