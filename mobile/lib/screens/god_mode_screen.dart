import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api.dart';
import '../services/auth.dart';

class GodModeScreen extends StatefulWidget {
  const GodModeScreen({super.key});
  @override
  State<GodModeScreen> createState() => _GodModeScreenState();
}

class _GodModeScreenState extends State<GodModeScreen> {
  final _starsCtrl = TextEditingController(text: '10000');
  final _userCtrl = TextEditingController();
  String? _status;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Developer Panel (God Mode)')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text('⚠️ Sandbox mode — available to everyone', style: TextStyle(color: Colors.orangeAccent)),
          const SizedBox(height: 24),
          const Text('Add Stars', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
          Row(
            children: [
              Expanded(child: TextField(controller: _starsCtrl, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: 'Amount'))),
              const SizedBox(width: 12),
              ElevatedButton(
                onPressed: () async {
                  final amount = int.tryParse(_starsCtrl.text) ?? 0;
                  final bal = await context.read<ApiService>().addStars(amount);
                  setState(() => _status = bal != null ? 'New balance: $bal ⭐' : 'Failed');
                  if (bal != null) {
                    final me = await context.read<ApiService>().getMe();
                    if (me != null) context.read<AuthService>().updateUser(me);
                  }
                },
                child: const Text('Add'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            icon: const Icon(Icons.star),
            label: const Text('Give myself Premium + Verified'),
            onPressed: () async {
              final ok = await context.read<ApiService>().setPremium(premium: true, verified: true, color: '#FF6B6B');
              setState(() => _status = ok ? 'Premium + Verified granted' : 'Failed');
              if (ok) {
                final me = await context.read<ApiService>().getMe();
                if (me != null) context.read<AuthService>().updateUser(me);
              }
            },
          ),
          const SizedBox(height: 16),
          const Text('Change username (any, including short)', style: TextStyle(fontWeight: FontWeight.bold)),
          Row(
            children: [
              Expanded(child: TextField(controller: _userCtrl, decoration: const InputDecoration(labelText: 'New username'))),
              const SizedBox(width: 12),
              ElevatedButton(
                onPressed: () async {
                  final ok = await context.read<ApiService>().changeUsername(_userCtrl.text.trim());
                  setState(() => _status = ok ? 'Username changed' : 'Failed (taken?)');
                  if (ok) {
                    final me = await context.read<ApiService>().getMe();
                    if (me != null) context.read<AuthService>().updateUser(me);
                  }
                },
                child: const Text('Set'),
              ),
            ],
          ),
          const SizedBox(height: 16),
          ElevatedButton.icon(
            icon: const Icon(Icons.card_giftcard),
            label: const Text('Grant first shop gift with custom #'),
            onPressed: () async {
              final shop = await context.read<ApiService>().getShopGifts();
              if (shop.isEmpty) {
                setState(() => _status = 'No gifts in shop');
                return;
              }
              final ok = await context.read<ApiService>().grantGift(shop[0]['id'], serial: 1);
              setState(() => _status = ok ? 'Gift granted #1' : 'Failed');
            },
          ),
          if (_status != null) ...[
            const SizedBox(height: 24),
            Text(_status!, style: const TextStyle(color: Colors.greenAccent, fontSize: 16)),
          ],
        ],
      ),
    );
  }
}
