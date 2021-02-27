class Kauthproxy < Formula
  desc "A kubectl plugin to forward a local port to a pod or service via authentication proxy"
  homepage "https://github.com/int128/kauthproxy"
  version "{{ env "VERSION" }}"

  on_macos do
    url "https://github.com/int128/kauthproxy/releases/download/{{ env "VERSION" }}/kauthproxy_darwin_amd64.zip"
    sha256 "{{ sha256 .darwin_amd64_archive }}"
  end
  on_linux do
    url "https://github.com/int128/kauthproxy/releases/download/{{ env "VERSION" }}/kauthproxy_linux_amd64.zip"
    sha256 "{{ sha256 .linux_amd64_archive }}"
  end

  def install
    bin.install "kauthproxy" => "kauthproxy"
    ln_s bin/"kauthproxy", bin/"kubectl-auth_proxy"
  end

  test do
    system "#{bin}/kauthproxy -h"
    system "#{bin}/kubectl-auth_proxy -h"
  end
end
