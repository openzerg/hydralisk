{ pkgs, lib, config, inputs, ... }:

let
  hydralisk-app = pkgs.buildGoModule {
    pname = "hydralisk";
    version = "1.0.0";
    src = ./.;
    vendorHash = "sha256-QSeG9TlxUIg6Wv5mYELAPfZcA+dbCxmOtqHfsD5JCYE=";
  };
  
  hydralisk-runtime = pkgs.buildEnv {
    name = "hydralisk-runtime";
    paths = [ hydralisk-app pkgs.bubblewrap pkgs.cacert pkgs.sqlite pkgs.bash ];
    pathsToLink = [ "/bin" "/etc" ];
  };
in
{
  languages.go.enable = true;

  packages = with pkgs; [
    buf
    protobuf
    sqlite
    bubblewrap
    inputs.nix2container.packages.${pkgs.system}.skopeo-nix2container
  ];

  processes = {
    hydralisk.exec = "go run . serve --host 0.0.0.0 --port 15317";
  };

  containers.hydralisk = {
    name = "hydralisk";
    copyToRoot = [ hydralisk-runtime ];
    startupCommand = "${hydralisk-app}/bin/hydralisk serve --host 0.0.0.0 --port 15317";
  };

  tasks = {
    "ci:build" = {
      exec = "go build -o hydralisk .";
    };
    "ci:test" = {
      exec = "go test ./...";
    };
    "container:copy-to-podman" = {
      exec = ''
        IMAGE=$(devenv container build hydralisk 2>&1 | tail -1)
        /nix/store/7v3jm6yyhk90gif8f2ra0702wwhpxz2d-skopeo-1.22.0/bin/skopeo copy nix:$IMAGE containers-storage:hydralisk:latest
        echo "Container copied to podman: hydralisk:latest"
      '';
    };
  };

  enterShell = ''
    echo "Hydralisk Development Environment"
    echo "Commands: go run . serve | go test ./... | go build"
    echo ""
    echo "Container commands:"
    echo "  devenv container build hydralisk                    - Build container"
  '';
}