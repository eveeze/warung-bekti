#!/usr/bin/env python3
"""
Test pricing engine on sample products.
"""
import re

def estimate_product_price(name, unit="pcs", category=""):
    """
    Estimates base_price and cost_price for Indonesian warung products.
    Returns (base_price, cost_price).
    """
    u_name = name.upper().strip()
    
    # 1. Parse Volume & Weight
    kg_match = re.search(r'(\d+(?:[.,]\d+)?)\s*(?:KG|KILOGRAM)', u_name)
    gram_match = re.search(r'(\d+(?:[.,]\d+)?)\s*(?:GR|G|GRAM)\b', u_name)
    liter_match = re.search(r'(\d+(?:[.,]\d+)?)\s*(?:LT|L|LITER)\b', u_name)
    ml_match = re.search(r'(\d+(?:[.,]\d+)?)\s*(?:ML|CC)\b', u_name)
    pcs_match = re.search(r'(\d+)\s*(?:PCS|BH|LEMBAR|TAB|KAPLET|BUTIR)\b', u_name)
    box_match = re.search(r'(\d+)\s*(?:S|PACK|BKS|\'S)\s*(?:BOX|PAK)?\b', u_name)
    meter_match = re.search(r'(\d+)\s*(?:M|METER|MTR)\b', u_name)

    kg = float(kg_match.group(1).replace(',', '.')) if kg_match else 0
    gram = float(gram_match.group(1).replace(',', '.')) if gram_match else 0
    liter = float(liter_match.group(1).replace(',', '.')) if liter_match else 0
    ml = float(ml_match.group(1).replace(',', '.')) if ml_match else 0
    meter = float(meter_match.group(1).replace(',', '.')) if meter_match else 0
    count = int(pcs_match.group(1)) if pcs_match else (int(box_match.group(1)) if box_match else 0)

    price = 0

    # ROKOK & TEMBAKAU
    if any(k in u_name for k in ["MILD", "SURYA", "GUDANG GARAM", "SAMPOERNA", "MARLBORO", "DJARUM", "MAGNUM", "DUNHILL", "ESSE", "CAMEL", "LA BOLD", "ROKOK", "TEMBAKAU", "PAPIR"]):
        if count == 12 or "12" in u_name:
            price = 22000
        elif count == 16 or "16" in u_name:
            price = 32000
        elif count == 20 or "20" in u_name:
            price = 42000
        elif "PAPIR" in u_name or "KLEMBAK" in u_name:
            price = 5000
        else:
            price = 28000

    # MIE INSTAN
    elif any(k in u_name for k in ["INDOMIE", "MIE SEDAAP", "SARIMI", "SUPERMI", "POP MIE", "GAGA", "INTERMIE", "SUKIMI", "BIHUN", "KWETIAW"]):
        if "POP MIE" in u_name or "CUP" in u_name:
            price = 6000
        elif "DUS" in u_name or "KARTON" in u_name:
            price = 125000
        elif "ISI 2" in u_name or "ISI2" in u_name or "JUMBO" in u_name:
            price = 4500
        else:
            price = 3500

    # SARDEN & KORNET
    elif any(k in u_name for k in ["SARDEN", "SARDINES", "KORNET", "TUNA KALENG"]):
        if gram >= 400 or "425" in u_name:
            price = 26000
        elif gram >= 300 or "340" in u_name:
            price = 36000
        elif "SACHET" in u_name or (gram and gram <= 60):
            price = 7500
        else:
            price = 12500

    # MINYAK GORENG
    elif any(k in u_name for k in ["MINYAK GORENG", "BIMOLI", "FILMA", "SANIA", "FORTUNE", "TROPICAL", "SUNCO", "SEDAAP MINYAK", "SOVIA", "KUNCI MAS", "MINYAKITA", "PALMA"]):
        if liter >= 5 or "5LT" in u_name or "5 L" in u_name:
            price = 90000
        elif liter >= 2 or "2LT" in u_name or "2 L" in u_name or "2000ML" in u_name:
            price = 36000
        elif liter >= 1 or "1LT" in u_name or "1 L" in u_name or "1000ML" in u_name:
            price = 18500
        else:
            price = 9500

    # BERAS
    elif any(k in u_name for k in ["BERAS"]):
        if kg >= 25 or "25KG" in u_name:
            price = 360000
        elif kg >= 10 or "10KG" in u_name:
            price = 150000
        elif kg >= 5 or "5KG" in u_name:
            price = 75000
        elif kg >= 2.5 or "2.5KG" in u_name:
            price = 38000
        else:
            price = 16000

    # GULA
    elif any(k in u_name for k in ["GULA", "GULAPASIR", "GULAKU"]):
        if kg >= 1 or "1KG" in u_name or gram >= 1000:
            price = 18000
        elif gram >= 500 or "500G" in u_name:
            price = 9500
        elif "MERAH" in u_name or "JAWA" in u_name or "AREN" in u_name:
            price = 15000
        else:
            price = 18000

    # TEPUNG
    elif any(k in u_name for k in ["TEPUNG", "TERIGU", "SEGITIGA BIRU", "CAKRA KEMBAR", "KUNCI BIRU", "MAIZENA", "TAPIOKA"]):
        if kg >= 1 or "1KG" in u_name or gram >= 1000 or "1000G" in u_name:
            price = 13500
        elif gram >= 500 or "500G" in u_name:
            price = 7500
        elif gram >= 200 or "250G" in u_name:
            price = 5000
        else:
            price = 12000

    # GARAM
    elif any(k in u_name for k in ["GARAM", "REFINA", "DOLPIN"]):
        if gram >= 500 or "500G" in u_name:
            price = 7500
        elif gram >= 200 or "250G" in u_name:
            price = 4000
        elif gram <= 100:
            price = 2500
        else:
            price = 4000

    # AIR MINERAL
    elif any(k in u_name for k in ["AIR MINERAL", "MINERAL WATER", "AQUA", "LE MINERALE", "CLEO", "VIT", "PRISTINE", "CLUB", "CHEERS", "ALTO", "FOREVER AIR"]):
        if liter >= 1.5 or "1500" in u_name or "1,5" in u_name:
            price = 6500
        elif ml >= 600 or "600" in u_name:
            price = 4000
        elif ml >= 300 or "330" in u_name:
            price = 3000
        elif "GALON" in u_name:
            price = 20000
        elif "CUP" in u_name or "GELAS" in u_name or ml <= 250:
            price = 1000
        else:
            price = 4000

    # KOPI & TEH
    elif any(k in u_name for k in ["KOPI", "NESCAFE", "KAPAL API", "GOOD DAY", "TORABIKA", "LUWAK", "TEH", "SOSRO", "SARIWANGI", "PUCUK"]):
        if gram >= 250 or "250G" in u_name:
            price = 25000
        elif gram >= 150 or "165G" in u_name or "200G" in u_name:
            price = 16000
        elif gram >= 60 or "65G" in u_name:
            price = 6500
        elif "BOTOL" in u_name or ml >= 300:
            price = 5000
        elif "RENCENG" in u_name or count >= 10:
            price = 15000
        elif "SACHET" in u_name or (count and count <= 5):
            price = 2000
        elif "KOTAK" in u_name:
            price = 4000
        else:
            price = 8500

    # SUSU
    elif any(k in u_name for k in ["SUSU", "DILAN", "ULTRA", "BEAR BRAND", "INDOMILK", "FRISIAN FLAG", "DANCOW", "MILO", "BENDERA", "CARNATION"]):
        if "1000ML" in u_name or liter >= 1 or "1L" in u_name:
            price = 20000
        elif "BEAR BRAND" in u_name:
            price = 11000
        elif "KALENG" in u_name or gram >= 350:
            price = 13500
        elif ml >= 200 or "250ML" in u_name or "200ML" in u_name:
            price = 6500
        elif ml >= 100 or "125ML" in u_name:
            price = 4000
        elif "SACHET" in u_name:
            price = 2500
        elif gram >= 800:
            price = 95000
        elif gram >= 400:
            price = 48000
        else:
            price = 9500

    # SABUN MANDI & SHAMPOO
    elif any(k in u_name for k in ["SABUN", "SHAMPOO", "SAMPO", "LIFEBUOY", "LUX", "DOVE", "PANTENE", "SUNSILK", "CLEAR", "REJOICE", "BIORE", "GIV", "NUVO", "DETTOL"]):
        if ml >= 700 or gram >= 700 or "900" in u_name:
            price = 45000
        elif ml >= 400 or gram >= 400 or "450ML" in u_name:
            price = 26000
        elif ml >= 200 or gram >= 200 or "250ML" in u_name:
            price = 17500
        elif ml >= 70 or "70ML" in u_name or "100ML" in u_name:
            price = 13500
        elif "BATANG" in u_name or "BAR" in u_name or (gram and gram <= 120):
            price = 4500
        elif "SACHET" in u_name or "RENCENG" in u_name:
            price = 12000
        else:
            price = 15000

    # DETERJEN & PEWANGI
    elif any(k in u_name for k in ["RINSO", "MOLTO", "DOWNY", "DAIA", "SO KLIN", "ATTACK", "BOOM", "JAZ1"]):
        if kg >= 1.5 or "1.8KG" in u_name or "1800G" in u_name:
            price = 38000
        elif kg >= 1 or gram >= 750 or "800G" in u_name or "750ML" in u_name or "700ML" in u_name:
            price = 19500
        elif gram >= 400 or "450G" in u_name:
            price = 10000
        elif gram >= 250 or "275G" in u_name:
            price = 7000
        elif "SACHET" in u_name:
            price = 1500
        else:
            price = 16000

    # SABUN CUCI PIRING
    elif any(k in u_name for k in ["SUNLIGHT", "MAMA LEMON", "EKONOMI"]):
        if ml >= 650 or gram >= 650 or "750ML" in u_name:
            price = 14500
        elif ml >= 350 or gram >= 350 or "400ML" in u_name:
            price = 8000
        elif ml >= 180 or gram >= 200:
            price = 5000
        elif "SACHET" in u_name or ml <= 120:
            price = 2500
        else:
            price = 8500

    # PASTA & SIKAT GIGI
    elif any(k in u_name for k in ["PEPSODENT", "CIPTADENT", "FORMULA", "CLOSE UP", "PASTA GIGI", "SIKAT GIGI"]):
        if gram >= 180 or "190G" in u_name:
            price = 16500
        elif gram >= 100 or "120G" in u_name:
            price = 11500
        elif gram >= 50 or "75G" in u_name:
            price = 6500
        elif "SIKAT" in u_name:
            price = 6000
        else:
            price = 9000

    # DIAPERS / POPOK
    elif any(k in u_name for k in ["MAMY", "MAMYPOKO", "SWEETY", "MERRIES", "BABY HAPPY", "FITTI", "POPOK"]):
        if count >= 40 or "L40" in u_name or "M40" in u_name or "XL" in u_name:
            price = 85000
        elif count >= 20 or "M20" in u_name or "L20" in u_name:
            price = 48000
        elif "RENCENG" in u_name or count <= 10:
            price = 18000
        else:
            price = 55000

    # MINYAK TELON & KAYU PUTIH
    elif any(k in u_name for k in ["TELON", "KAYU PUTIH", "MY BABY", "CAP LANG", "KONICARE"]):
        if ml >= 120 or "120ML" in u_name or "150ML" in u_name:
            price = 42000
        elif ml >= 50 or "60ML" in u_name:
            price = 24000
        elif ml >= 25 or "30ML" in u_name:
            price = 14000
        else:
            price = 22000

    # OBAT P3K & FARMASI
    elif any(k in u_name for k in ["PARACETAMOL", "BODREX", "PARAMEX", "PROMAG", "PANADOL", "MIXAGRIP", "KOMIX", "TOLAK ANGIN", "ANTANGIN", "COMBANTRIN", "BETADINE", "HANSAPLAST", "VISINE", "SAKATONIK", "MEXTRIL"]):
        if "310ML" in u_name or "300ML" in u_name:
            price = 38000
        elif "200ML" in u_name:
            price = 28000
        elif "100ML" in u_name or "125ML" in u_name:
            price = 22000
        elif "COMBANTRIN" in u_name:
            price = 24000
        elif "VISINE" in u_name:
            price = 18000
        elif "KAPLET" in u_name or "TABLET" in u_name or "STRIP" in u_name:
            price = 6500
        elif "BOX" in u_name or count >= 10:
            price = 25000
        elif "SACHET" in u_name:
            price = 3500
        else:
            price = 12000

    # BUKU & ALAT TULIS
    elif any(k in u_name for k in ["BUKU", "SIDU", "KIKY", "PULPEN", "PENSIL", "SPIDOL", "PENGHAPUS", "LAKBAN", "ISOLASI", "P/CASE"]):
        if "PAK" in u_name or "DUS" in u_name or "10" in u_name and "BUKU" in u_name:
            price = 38000
        elif "BUKU" in u_name:
            if "58" in u_name:
                price = 5500
            elif "38" in u_name:
                price = 4000
            elif "KAMPUS" in u_name or "BOXY" in u_name:
                price = 7500
            else:
                price = 5000
        elif "LAKBAN" in u_name:
            price = 12000
        elif "PULPEN" in u_name or "PENSIL" in u_name:
            price = 3500
        elif "SPIDOL" in u_name:
            price = 8500
        else:
            price = 5000

    # KABEL & LISTRIK
    elif any(k in u_name for k in ["KABEL", "STOP KONTAK", "SOKET", "STEKER", "LOYAL", "HIKARI", "CORDMAN", "LAMPU", "LED"]):
        if meter >= 20 or "20 M" in u_name:
            price = 75000
        elif meter >= 15 or "15 M" in u_name:
            price = 58000
        elif meter >= 10 or "10 M" in u_name:
            price = 45000
        elif "STOP KONTAK" in u_name or "SOKET" in u_name:
            price = 22000
        elif "STEKER" in u_name or "COLOKAN" in u_name:
            price = 8500
        elif "LAMPU" in u_name or "LED" in u_name:
            price = 25000
        else:
            price = 25000

    # BISKUIT, SNACK, COKLAT
    elif any(k in u_name for k in ["BISKUIT", "WAFER", "CHITATO", "QTELA", "TANGO", "NABATI", "OREO", "ROMA", "MALKIST", "SILVERQUEEN", "BENG-BENG", "COKLAT", "PERMEN", "PARAGO"]):
        if "KALENG" in u_name or gram >= 500:
            price = 65000
        elif "SILVERQUEEN" in u_name:
            price = 16000
        elif gram >= 200 or "250G" in u_name or "300G" in u_name:
            price = 15000
        elif gram >= 100 or "120G" in u_name:
            price = 9500
        elif gram >= 50 or "68G" in u_name:
            price = 6500
        elif "BENG-BENG" in u_name:
            price = 2500
        elif "PERMEN" in u_name or "LOLIPO" in u_name:
            price = 1000
        else:
            price = 5000

    # PLASTIK / WARE (HAWAII, SHINPO)
    elif any(k in u_name for k in ["HAWAII", "SHINPO", "LION STAR", "BASKOM", "EMBER", "TOPLES", "KERANJANG"]):
        price = 18000

    # KECAP, SAUS, BUMBU DAPUR
    elif any(k in u_name for k in ["KECAP", "SAOS", "SAUS", "BANGO", "LADAKU", "DESAKU", "MASAKO", "ROYCO", "SASA", "AJINOMOTO", "KARA", "CUKA"]):
        if ml >= 500 or gram >= 500:
            price = 23000
        elif ml >= 200 or gram >= 200:
            price = 10500
        elif ml >= 100 or gram >= 100:
            price = 6500
        elif "SACHET" in u_name or "RENCENG" in u_name:
            price = 5500
        else:
            price = 6000

    # COSMETICS & SKINCARE (VIVA, WARDAH, PONDS, PIXY, NIVEA, LATULIPE)
    elif any(k in u_name for k in ["VIVA", "WARDAH", "PONDS", "PIXY", "NIVEA", "LATULIPE", "GARNIER", "OVALE", "BEDAK", "LIPSTIK", "FACIAL"]):
        if "FACIAL" in u_name or "FOAM" in u_name or gram >= 100:
            price = 32000
        elif "50G" in u_name or "50ML" in u_name:
            price = 18000
        elif ml >= 100:
            price = 24000
        elif "BEDAK" in u_name:
            price = 22000
        else:
            price = 20000

    # Fallback based on category
    if price == 0:
        if "Rokok" in category:
            price = 28000
        elif "Minuman" in category:
            price = 5000
        elif "Makanan & Mie" in category:
            price = 6000
        elif "Sembako" in category:
            price = 15000
        elif "Bumbu" in category:
            price = 6000
        elif "Pembersih" in category:
            price = 12000
        elif "Perawatan" in category:
            price = 15000
        elif "Bayi" in category:
            price = 35000
        elif "Obat" in category:
            price = 12000
        elif "Alat Tulis" in category:
            price = 5000
        elif "Sandang" in category:
            price = 18000
        else:
            price = 10000

    # Ensure commercial rounding to nearest 500
    price = int(round(price / 500.0) * 500)
    if price <= 0:
        price = 1000

    # Cost price is ~82% of base price (18% gross margin), rounded to nearest 100
    cost_price = int(round((price * 0.82) / 100.0) * 100)
    if cost_price >= price:
        cost_price = price - 500
    if cost_price <= 0:
        cost_price = max(500, price - 500)

    return price, cost_price

if __name__ == "__main__":
    test_items = [
        ("O VISINE 6ML EXTRA", "Obat & Kesehatan"),
        ("COMBANTRIN 2’S", "Obat & Kesehatan"),
        ("O SAKATONIK LIVER 310ML", "Obat & Kesehatan"),
        ("INDOMIE GORENG", "Makanan & Mie Instan"),
        ("SARIMI ISI2 BASO SAPI", "Makanan & Mie Instan"),
        ("GULAPASIR RAJA GULA", "Sembako & Bahan Pokok"),
        ("GARAM REFINA 250G", "Sembako & Bahan Pokok"),
        ("GARAM REFINA 500G", "Sembako & Bahan Pokok"),
        ("CAKRA KEMBAR TEPUNG TERIGUPCK 1000g", "Sembako & Bahan Pokok"),
        ("SEGITIGA BIRU 1KG TERIGU", "Sembako & Bahan Pokok"),
        ("PALMA 5LT", "Sembako & Bahan Pokok"),
        ("FOREVER AIR MINERAL 1500ML", "Minuman"),
        ("ALTO MINERAL WATER 600ML", "Minuman"),
        ("PARAGO COKLAT SUSU", "Minuman"),
        ("KOPI BUBUK BERONTOSENO 250G", "Minuman"),
        ("CLAS MILD 16", "Rokok & Tembakau"),
        ("CLAS MILD 12", "Rokok & Tembakau"),
        ("KABEL BOX LOYAL 10 METER", "Alat Tulis & Kantor"),
        ("KABEL BOX LOYAL 15 METER FULL", "Alat Tulis & Kantor"),
        ("DORIS BEDAK PMBSI KUTU 100 GR", "Perawatan Tubuh & Kebersihan"),
        ("RINSO MOLTO ROYAL GOLD LIQUID", "Pembersih & Rumah Tangga"),
        ("SUNLIGHT LIME POUCH 750ML", "Pembersih & Rumah Tangga"),
        ("MAMYPOKO PANTS STANDARD M34", "Perlengkapan Bayi & Anak"),
    ]
    for name, cat in test_items:
        bp, cp = estimate_product_price(name, "pcs", cat)
        margin = ((bp - cp) / bp) * 100
        print(f"{name:<35} | {cat:<25} -> Jual: Rp {bp:>7,d} | Pokok: Rp {cp:>7,d} (Margin {margin:.1f}%)")
