using System;
using System.IO;
using Microsoft.Maui.Controls;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public partial class ProfilePage : ContentPage
    {
        private const long MaxAvatarBytes = 5 * 1024 * 1024;
        private readonly ApiService _api = new();

        public ProfilePage()
        {
            InitializeComponent();
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadProfileAsync();
        }

        private async System.Threading.Tasks.Task LoadProfileAsync()
        {
            var user = await _api.GetCurrentUserAsync();
            if (user == null)
            {
                return;
            }

            NameLabel.Text = user.Name;
            PhoneLabel.Text = user.Phone;
            IdLabel.Text = $"ID: {user.Id}";
            AgeLabel.Text = $"Возраст: {user.Age}";
            CreatedLabel.Text = user.CreatedAt == default
                ? "Дата регистрации: не указана"
                : $"Дата регистрации: {user.CreatedAt:dd.MM.yyyy HH:mm}";

            var avatarBytes = await _api.GetAvatarAsync();
            if (avatarBytes != null && avatarBytes.Length > 0)
            {
                AvatarImage.Source = ImageSource.FromStream(() => new MemoryStream(avatarBytes));
            }
        }

        private async void OnAvatarClicked(object sender, EventArgs e)
        {
            try
            {
                var result = await FilePicker.PickAsync(new PickOptions
                {
                    PickerTitle = "Выберите JPG или PNG"
                });
                if (result == null) return;

                var ext = Path.GetExtension(result.FileName).ToLowerInvariant();
                if (ext != ".jpg" && ext != ".jpeg" && ext != ".png")
                {
                    await DisplayAlert("Ошибка", "Аватар должен быть в формате JPG или PNG.", "OK");
                    return;
                }

                await using var stream = await result.OpenReadAsync();
                using var memory = new MemoryStream();
                await stream.CopyToAsync(memory);
                var bytes = memory.ToArray();
                if (bytes.Length > MaxAvatarBytes)
                {
                    await DisplayAlert("Ошибка", "Размер аватара не должен превышать 5 MB.", "OK");
                    return;
                }

                var ok = await _api.UploadAvatarAsync(result.FileName, bytes);
                if (!ok)
                {
                    await DisplayAlert("Ошибка", "Сервер отклонил аватар. Проверьте формат, размер и пропорции.", "OK");
                    return;
                }

                AvatarImage.Source = ImageSource.FromStream(() => new MemoryStream(bytes));
                await DisplayAlert("Готово", "Аватар обновлён.", "OK");
            }
            catch (Exception ex)
            {
                await DisplayAlert("Ошибка", ex.Message, "OK");
            }
        }
    }
}
