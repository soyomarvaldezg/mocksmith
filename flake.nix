{
  description = "Mocksmith - generate realistic fake data for testing and demos";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      version = self.shortRev or "dev";
      commit = self.rev or "unknown";
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          mocksmith = pkgs.buildGoModule {
            pname = "mocksmith";
            inherit version;
            src = self;
            vendorHash = "sha256-avzxiRO+ZLGCaKAS34FkueepUpSvy5hBONLe2vhpf3Y=";
            subPackages = [ "." ];
            ldflags = [
              "-s"
              "-w"
              "-X github.com/soyomarvaldezg/mocksmith/cmd.version=${version}"
              "-X github.com/soyomarvaldezg/mocksmith/cmd.commit=${commit}"
              "-X github.com/soyomarvaldezg/mocksmith/cmd.date=unknown"
            ];
            meta = with pkgs.lib; {
              description = "Generate realistic fake data for testing and demos";
              homepage = "https://github.com/soyomarvaldezg/mocksmith";
              license = licenses.mit;
              mainProgram = "mocksmith";
            };
          };
        in
        {
          inherit mocksmith;
          default = mocksmith;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.just
              pkgs.goreleaser
              pkgs.cocogitto
            ];
          };
        }
      );
    };
}
