import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;

class AuthService extends ChangeNotifier {
  String? _token;
  Map<String, dynamic>? _user;
  bool isLoading = true;

  String? get token => _token;
  Map<String, dynamic>? get user => _user;
  bool get isLoggedIn => _token != null;

  // Change this to your server address
  static const String baseUrl = 'http://10.0.2.2:8080'; // Android emulator
  // static const String baseUrl = 'http://localhost:8080'; // Windows / Web
  // static const String baseUrl = 'http://YOUR_SERVER_IP:8080';

  AuthService() {
    _load();
  }

  Future<void> _load() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('token');
    final userJson = prefs.getString('user');
    if (userJson != null) {
      _user = jsonDecode(userJson);
    }
    isLoading = false;
    notifyListeners();
  }

  Future<String?> register(String username, String password, {String email = ''}) async {
    try {
      final res = await http.post(
        Uri.parse('$baseUrl/api/auth/register'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'username': username, 'password': password, 'email': email}),
      );
      final data = jsonDecode(res.body);
      if (res.statusCode == 201) {
        await _save(data['access_token'], data['user']);
        return null;
      }
      return data['error'] ?? 'Registration failed';
    } catch (e) {
      return e.toString();
    }
  }

  Future<String?> login(String username, String password) async {
    try {
      final res = await http.post(
        Uri.parse('$baseUrl/api/auth/login'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'username': username, 'password': password}),
      );
      final data = jsonDecode(res.body);
      if (res.statusCode == 200) {
        await _save(data['access_token'], data['user']);
        return null;
      }
      return data['error'] ?? 'Login failed';
    } catch (e) {
      return e.toString();
    }
  }

  Future<void> _save(String token, Map<String, dynamic> user) async {
    _token = token;
    _user = user;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('token', token);
    await prefs.setString('user', jsonEncode(user));
    notifyListeners();
  }

  Future<void> logout() async {
    _token = null;
    _user = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.clear();
    notifyListeners();
  }

  void updateUser(Map<String, dynamic> user) {
    _user = user;
    SharedPreferences.getInstance().then((p) => p.setString('user', jsonEncode(user)));
    notifyListeners();
  }
}
