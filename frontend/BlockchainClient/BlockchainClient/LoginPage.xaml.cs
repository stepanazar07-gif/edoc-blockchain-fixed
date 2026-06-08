using System;
using Microsoft.Maui.Controls;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public partial class LoginPage : ContentPage
    {
        private readonly ApiService _api = new();

        public LoginPage()
        {
            InitializeComponent();
        }

        private async void OnLoginClicked(object sender, EventArgs e)
        {
            MessageLabel.Text = "";

            var phone = PhoneEntry.Text?.Trim() ?? "";
            var password = PasswordEntry.Text?.Trim() ?? "";
            if (string.IsNullOrWhiteSpace(phone) || string.IsNullOrWhiteSpace(password))
            {
                MessageLabel.Text = "Введите телефон и пароль.";
                return;
            }

            var token = await _api.LoginAsync(phone, password);
            if (token == null)
            {
                MessageLabel.Text = string.IsNullOrWhiteSpace(_api.LastError)
                    ? "Неверный телефон или пароль."
                    : $"Ошибка входа: {_api.LastError}";
                return;
            }

            await SecureStorage.SetAsync("auth_token", token);
            SessionNavigator.ShowMain();
        }

        private async void OnRegisterClicked(object sender, EventArgs e)
        {
            MessageLabel.Text = "";

            var name = NameEntry.Text?.Trim() ?? "";
            var phone = PhoneEntry.Text?.Trim() ?? "";
            var password = PasswordEntry.Text?.Trim() ?? "";

            if (!int.TryParse(AgeEntry.Text?.Trim(), out var age))
            {
                MessageLabel.Text = "Возраст должен быть числом.";
                return;
            }
            if (string.IsNullOrWhiteSpace(name) || string.IsNullOrWhiteSpace(phone) || string.IsNullOrWhiteSpace(password))
            {
                MessageLabel.Text = "Для регистрации заполните имя, возраст, телефон и пароль.";
                return;
            }
            if (password.Length < 6)
            {
                MessageLabel.Text = "Пароль должен быть не короче 6 символов.";
                return;
            }

            var token = await _api.RegisterAsync(name, age, phone, password);
            if (token == null)
            {
                MessageLabel.Text = string.IsNullOrWhiteSpace(_api.LastError)
                    ? "Не удалось зарегистрироваться. Проверьте данные или уникальность телефона."
                    : $"Не удалось зарегистрироваться: {_api.LastError}";
                return;
            }

            await SecureStorage.SetAsync("auth_token", token);
            SessionNavigator.ShowMain();
        }
    }
}
