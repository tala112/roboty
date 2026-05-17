# -*- mode: ruby -*-
# vi: set ft=ruby :
#
# Roboty VM Test Matrix — cross-platform safety testing.
# Usage:
#   vagrant up win11        # Windows 11 (P0)
#   vagrant up ubuntu       # Ubuntu 24.04 GNOME Wayland (P0)
#   vagrant ssh ubuntu -c 'cd /vagrant && make test-all'
#
# macOS requires a separate macOS host (Vagrant doesn't support macOS guests).
# Use GitHub Actions macos-latest for macOS testing.

Vagrant.configure("2") do |config|
  config.vm.provider "virtualbox" do |vb|
    vb.memory = 4096
    vb.cpus = 2
  end

  # === Windows 11 24H2 (P0) ===
  config.vm.define "win11", primary: true do |win11|
    win11.vm.box = "gusztavvargadr/windows-11-24h2-enterprise"
    win11.vm.provider "virtualbox" do |vb|
      vb.gui = true
      vb.customize ["modifyvm", :id, "--clipboard", "bidirectional"]
    end
    win11.vm.provision "shell", inline: <<-SHELL
      # Install Go
      $url = "https://go.dev/dl/go1.26.3.windows-amd64.msi"
      $msi = "$env:TEMP\go.msi"
      [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
      Invoke-WebRequest -Uri $url -OutFile $msi
      Start-Process msiexec.exe -Wait -ArgumentList "/i $msi /quiet"
      $env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine")
      [Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Go\bin", "Machine")
      # Run tests
      cd C:\vagrant
      go test -count=1 -run TestCritical ./internal/modes/
      go test -fuzz=FuzzNormalizeKillExec -fuzztime=10s ./internal/modes/
    SHELL
  end

  # === Ubuntu 24.04 GNOME Wayland (P0) ===
  config.vm.define "ubuntu", primary: true do |ubuntu|
    ubuntu.vm.box = "ubuntu/noble64"
    ubuntu.vm.provider "virtualbox" do |vb|
      vb.gui = true
      vb.customize ["modifyvm", :id, "--graphicscontroller", "vmsvga"]
    end
    ubuntu.vm.provision "shell", inline: <<-SHELL
      apt-get update
      apt-get install -y golang-go gnome-session wayland
      export CGO_ENABLED=1
      cd /vagrant
      go test -count=1 -run TestCritical ./internal/modes/
      go test -fuzz=FuzzNormalizeKillExec -fuzztime=10s ./internal/modes/
    SHELL
  end

  # === Windows 10 22H2 (P1) ===
  config.vm.define "win10" do |win10|
    win10.vm.box = "gusztavvargadr/windows-10-22h2-enterprise"
    win10.vm.provider "virtualbox" do |vb|
      vb.gui = true
    end
    win10.vm.provision "shell", inline: <<-SHELL
      $url = "https://go.dev/dl/go1.26.3.windows-amd64.msi"
      $msi = "$env:TEMP\go.msi"
      [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
      Invoke-WebRequest -Uri $url -OutFile $msi
      Start-Process msiexec.exe -Wait -ArgumentList "/i $msi /quiet"
      cd C:\vagrant
      go test -count=1 -run TestCritical ./internal/modes/
    SHELL
  end
end
