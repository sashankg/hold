apple:
	cd packet_handler; cargo swift package --name PacketHandler --platforms macos --platforms ios

ios:
	cd packet_handler; cargo swift package --name PacketHandler --platforms ios

macos:
	cd packet_handler; cargo swift package --name PacketHandler --platforms macos
