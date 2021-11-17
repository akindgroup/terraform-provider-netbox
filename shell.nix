with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    gnumake
    go
  ];
  shellHook = ''
    export NETBOX_HOST="https://netbox.academicwork.net"
  '';
}