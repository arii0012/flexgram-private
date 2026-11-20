import 'dart:convert';
import 'package:http/http.dart' as http;
import 'auth.dart';

class ApiService {
  final AuthService auth;
  ApiService(this.auth);

  String get base => AuthService.baseUrl;

  Map<String, String> get _headers => {
        'Content-Type': 'application/json',
        if (auth.token != null) 'Authorization': 'Bearer ${auth.token}',
      };

  Future<List<dynamic>> getChats() async {
    final res = await http.get(Uri.parse('$base/api/chats'), headers: _headers);
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['chats'] ?? [];
    }
    return [];
  }

  Future<List<dynamic>> getMessages(String chatId) async {
    final res = await http.get(Uri.parse('$base/api/chats/$chatId/messages'), headers: _headers);
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['messages'] ?? [];
    }
    return [];
  }

  Future<Map<String, dynamic>?> sendMessage(String chatId, String content) async {
    final res = await http.post(
      Uri.parse('$base/api/chats/$chatId/messages'),
      headers: _headers,
      body: jsonEncode({'content': content}),
    );
    if (res.statusCode == 201) {
      return jsonDecode(res.body)['message'];
    }
    return null;
  }

  Future<String?> createDirectChat(String username) async {
    final res = await http.post(
      Uri.parse('$base/api/chats'),
      headers: _headers,
      body: jsonEncode({'type': 'direct', 'username': username}),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode == 201 || res.statusCode == 200) {
      return data['chat_id'];
    }
    return null;
  }

  Future<List<dynamic>> searchUsers(String q) async {
    final res = await http.get(Uri.parse('$base/api/users/search?q=$q'), headers: _headers);
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['users'] ?? [];
    }
    return [];
  }

  Future<Map<String, dynamic>?> getMe() async {
    final res = await http.get(Uri.parse('$base/api/users/me'), headers: _headers);
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['user'];
    }
    return null;
  }

  Future<List<dynamic>> getShopGifts() async {
    final res = await http.get(Uri.parse('$base/api/gifts/shop'), headers: _headers);
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['gifts'] ?? [];
    }
    return [];
  }

  Future<List<dynamic>> getMyGifts() async {
    final res = await http.get(Uri.parse('$base/api/gifts/my'), headers: _headers);
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['gifts'] ?? [];
    }
    return [];
  }

  Future<bool> buyGift(String giftId) async {
    final res = await http.post(
      Uri.parse('$base/api/gifts/buy'),
      headers: _headers,
      body: jsonEncode({'gift_id': giftId}),
    );
    return res.statusCode == 200;
  }

  // God Mode
  Future<int?> addStars(int amount) async {
    final res = await http.post(
      Uri.parse('$base/api/god/stars'),
      headers: _headers,
      body: jsonEncode({'amount': amount}),
    );
    if (res.statusCode == 200) {
      return jsonDecode(res.body)['balance'];
    }
    return null;
  }

  Future<bool> setPremium({bool premium = true, bool verified = true, String color = '#2AABEE'}) async {
    final res = await http.post(
      Uri.parse('$base/api/god/premium'),
      headers: _headers,
      body: jsonEncode({'premium': premium, 'verified': verified, 'color': color}),
    );
    return res.statusCode == 200;
  }

  Future<bool> grantGift(String giftId, {int serial = 0}) async {
    final res = await http.post(
      Uri.parse('$base/api/god/gift'),
      headers: _headers,
      body: jsonEncode({'gift_id': giftId, 'serial_number': serial}),
    );
    return res.statusCode == 200;
  }

  Future<bool> changeUsername(String username) async {
    final res = await http.post(
      Uri.parse('$base/api/god/username'),
      headers: _headers,
      body: jsonEncode({'username': username}),
    );
    return res.statusCode == 200;
  }
}
