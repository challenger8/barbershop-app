(1, 3, 'Traditional Beard Sculpting', 'Artisanal beard trimming using traditional Italian techniques', 30.00, 40.00, 'USD', NULL, NULL, 25, 35, 10, 2, 30, '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]', '{}', false, NULL, 'Come with at least 2 weeks of beard growth', 'Apply provided beard oil daily', 18, 70, false, NULL, NULL, '["https://images.unsplash.com/photo-1621605815971-fbc98d665033?w=400"]', '["https://beard-before-after-1.jpg"]', 156, 4680.00, 4.7, 67, 3.2, 94.1, 82.4, 12, 360.00, 89.2, 0.7, false, NULL, NULL, NULL, false, 2, 'Includes premium beard oil treatment', true, NULL, NULL, NOW() - INTERVAL '20 days', NOW()),

(1, 4, 'Authentic Italian Hot Towel Shave', 'Traditional straight razor shave with family recipe pre-shave oils', 65.00, 75.00, 'USD', NULL, NULL, 50, 60, 20, 4, 14, '["tuesday", "wednesday", "thursday", "friday", "saturday"]', '{}', true, 15, 'No caffeine 2 hours before appointment', 'Avoid sun exposure for 24 hours', 18, 80, false, NULL, NULL, '["https://images.unsplash.com/photo-1585747860715-2ba37e788b70?w=400"]', '[]', 89, 5785.00, 4.9, 42, 1.1, 98.8, 71.4, 7, 455.00, 78.4, 0.9, false, NULL, NULL, NULL, true, 3, 'Premium service, requires consultation', true, NULL, NULL, NOW() - INTERVAL '20 days', NOW()),

-- Maria's Services
(2, 2, 'Trendy Fade Design', 'Modern fade cuts with creative design elements and styling', 50.00, 65.00, 'USD', 45.00, NOW() + INTERVAL '14 days', 45, 60, 15, 4, 45, '["tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"]', '{}', true, 20, 'Bring inspiration photos if you have specific ideas', 'Use sulfate-free shampoo for color protection', 14, 50, false, NULL, NULL, '["https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=400"]', '["https://fade-before-after-1.jpg", "https://fade-before-after-2.jpg"]', 187, 9350.00, 4.6, 78, 4.8, 91.2, 65.8, 14, 700.00, 92.8, 0.8, true, 'New client special: $5 off first fade!', NOW(), NOW() + INTERVAL '14 days', true, 1, 'Consultation included in price', true, NULL, NULL, NOW() - INTERVAL '25 days', NOW()),

(2, 6, 'Luxury Hair Treatment & Style', 'Deep conditioning treatment with professional styling', 55.00, 70.00, 'USD', NULL, NULL, 75, 90, 15, 4, 45, '["tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"]', '{}', false, NULL, 'Arrive with unwashed hair for best treatment results', 'Avoid heat styling for 48 hours', 12, 80, false, NULL, NULL, '["https://images.unsplash.com/photo-1562004760-acb5f2f1dfef?w=400"]', '[]', 92, 5060.00, 4.4, 34, 6.5, 88.9, 58.7, 8, 440.00, 75.3, 0.6, false, NULL, NULL, NULL, false, 2, 'Perfect for special occasions', true, NULL, NULL, NOW() - INTERVAL '25 days', NOW()),

-- David's Services
(3, 1, 'Precision Business Cut', 'Military-precision haircut perfect for professionals', 40.00, 50.00, 'USD', NULL, NULL, 30, 40, 10, 3, 60, '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]', '{}', false, NULL, 'Please arrive 5 minutes early', 'Style as usual, cut will hold shape for 4-6 weeks', 16, 70, false, NULL, NULL, '["https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=400"]', '["https://precision-before-after-1.jpg"]', 298, 11920.00, 4.9, 124, 1.7, 97.8, 84.2, 22, 880.00, 96.8, 0.9, false, NULL, NULL, NULL, true, 1, 'Satisfaction guaranteed or your money back', true, NULL, NULL, NOW() - INTERVAL '18 days', NOW()),

(3, 2, 'Modern Fade Mastery', 'Expert fade cutting with Asian hair specialization', 55.00, 70.00, 'USD', NULL, NULL, 45, 55, 15, 3, 60, '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]', '{}', true, 15, 'Bring reference photos for best results', 'Use recommended styling products for maintenance', 14, 45, false, NULL, NULL, '["https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=400"]', '["https://asian-fade-before-after-1.jpg", "https://asian-fade-before-after-2.jpg"]', 267, 14685.00, 4.8, 98, 2.2, 96.1, 79.3, 19, 1045.00, 94.7, 0.8, false, NULL, NULL, NULL, true, 2, 'Specializing in Asian hair textures', true, NULL, NULL, NOW() - INTERVAL '18 days', NOW()),

(3, 5, 'Kids Precision Cut', 'Patient, gentle haircuts for children with precision results', 25.00, 30.00, 'USD', NULL, NULL, 20, 25, 5, 2, 30, '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]', '{}', false, NULL, 'Bring favorite toy or tablet for entertainment', 'Regular trims every 6-8 weeks recommended', 3, 17, false, NULL, NULL, '["https://images.unsplash.com/photo-1564463489817-3f6eccb47b89?w=400"]', '[]', 145, 3625.00, 4.7, 52, 3.4, 92.3, 71.0, 11, 275.00, 85.7, 0.6, false, NULL, NULL, NULL, false, 3, 'Kid-friendly environment with games', true, NULL, NULL, NOW() - INTERVAL '18 days', NOW());

-- =============================================================================
-- 6. TIME SLOTS SEED DATA
-- =============================================================================

INSERT INTO time_slots (barber_id, start_time, end_time, duration_minutes, is_available, slot_type, base_price, dynamic_price, discount_percentage, service_id, max_customers, min_advance_notice_hours, notes, special_requirements, created_at, updated_at, created_by) VALUES

-- Tony's Time Slots (next 7 days)
(1, NOW() + INTERVAL '1 day 9:00:00', NOW() + INTERVAL '1 day 9:45:00', 45, true, 'regular', 45.00, NULL, 0, 1, 1, 2, 'Morning slot', '{}', NOW(), NOW(), 1),
(1, NOW() + INTERVAL '1 day 10:00:00', NOW() + INTERVAL '1 day 10:45:00', 45, true, 'regular', 45.00, NULL, 0, 1, 1, 2, NULL, '{}', NOW(), NOW(), 1),
(1, NOW() + INTERVAL '1 day 11:00:00', NOW() + INTERVAL '1 day 11:45:00', 45, true, 'regular', 45.00, NULL, 0, 1, 1, 2, NULL, '{}', NOW(), NOW(), 1),
(1, NOW() + INTERVAL '1 day 14:00:00', NOW() + INTERVAL '1 day 14:35:00', 35, true, 'regular', 30.00, NULL, 0, 3, 1, 2, 'Afternoon beard trim', '{}', NOW(), NOW(), 1),
(1, NOW() + INTERVAL '1 day 15:00:00', NOW() + INTERVAL '1 day 16:00:00', 60, true, 'premium', 65.00, NULL, 0, 4, 1, 4, 'Hot towel shave - consultation required', '{"consultation": true}', NOW(), NOW(), 1),

(1, NOW() + INTERVAL '2 days 9:00:00', NOW() + INTERVAL '2 days 9:45:00', 45, true, 'regular', 45.00, NULL, 0, 1, 1, 2, NULL, '{}', NOW(), NOW(), 1),
(1, NOW() + INTERVAL '2 days 10:00:00', NOW() + INTERVAL '2 days 10:45:00', 45, true, 'regular', 45.00, NULL, 0, 1, 1, 2, NULL, '{}', NOW(), NOW(), 1),
(1, NOW() + INTERVAL '2 days 14:00:00', NOW() + INTERVAL '2 days 14:35:00', 35, true, 'regular', 30.00, NULL, 0, 3, 1, 2, NULL, '{}', NOW(), NOW(), 1),

-- Maria's Time Slots
(2, NOW() + INTERVAL '1 day 10:00:00', NOW() + INTERVAL '1 day 11:00:00', 60, true, 'regular', 50.00, NULL, 10, 2, 1, 4, 'New client special available', '{}', NOW(), NOW(), 1),
(2, NOW() + INTERVAL '1 day 11:30:00', NOW() + INTERVAL '1 day 12:30:00', 60, true, 'regular', 50.00, NULL, 10, 2, 1, 4, NULL, '{}', NOW(), NOW(), 1),
(2, NOW() + INTERVAL '1 day 14:00:00', NOW() + INTERVAL '1 day 15:30:00', 90, true, 'premium', 55.00, NULL, 0, 6, 1, 4, 'Luxury treatment', '{}', NOW(), NOW(), 1),
(2, NOW() + INTERVAL '1 day 16:00:00', NOW() + INTERVAL '1 day 17:00:00', 60, true, 'regular', 50.00, NULL, 10, 2, 1, 4, NULL, '{}', NOW(), NOW(), 1),

(2, NOW() + INTERVAL '2 days 10:00:00', NOW() + INTERVAL '2 days 11:00:00', 60, true, 'regular', 50.00, NULL, 10, 2, 1, 4, NULL, '{}', NOW(), NOW(), 1),
(2, NOW() + INTERVAL '2 days 14:00:00', NOW() + INTERVAL '2 days 15:30:00', 90, true, 'premium', 55.00, NULL, 0, 6, 1, 4, NULL, '{}', NOW(), NOW(), 1),

-- David's Time Slots
(3, NOW() + INTERVAL '1 day 10:00:00', NOW() + INTERVAL '1 day 10:40:00', 40, true, 'regular', 40.00, NULL, 0, 1, 1, 3, NULL, '{}', NOW(), NOW(), 1),
(3, NOW() + INTERVAL '1 day 11:00:00', NOW() + INTERVAL '1 day 11:55:00', 55, true, 'regular', 55.00, NULL, 0, 2, 1, 3, 'Consultation included', '{"consultation": true}', NOW(), NOW(), 1),
(3, NOW() + INTERVAL '1 day 14:00:00', NOW() + INTERVAL '1 day 14:40:00', 40, true, 'regular', 40.00, NULL, 0, 1, 1, 3, NULL, '{}', NOW(), NOW(), 1),
(3, NOW() + INTERVAL '1 day 15:00:00', NOW() + INTERVAL '1 day 15:25:00', 25, true, 'regular', 25.00, NULL, 0, 5, 1, 2, 'Kids cut', '{}', NOW(), NOW(), 1),
(3, NOW() + INTERVAL '1 day 16:00:00', NOW() + INTERVAL '1 day 16:55:00', 55, true, 'regular', 55.00, NULL, 0, 2, 1, 3, NULL, '{"consultation": true}', NOW(), NOW(), 1),

(3, NOW() + INTERVAL '2 days 10:00:00', NOW() + INTERVAL '2 days 10:40:00', 40, true, 'regular', 40.00, NULL, 0, 1, 1, 3, NULL, '{}', NOW(), NOW(), 1),
(3, NOW() + INTERVAL '2 days 11:00:00', NOW() + INTERVAL '2 days 11:55:00', 55, true, 'regular', 55.00, NULL, 0, 2, 1, 3, NULL, '{"consultation": true}', NOW(), NOW(), 1);

-- =============================================================================
-- 7. BOOKINGS SEED DATA
-- =============================================================================

INSERT INTO bookings (uuid, booking_number, customer_id, barber_id, time_slot_id, service_name, service_category, estimated_duration_minutes, customer_name, customer_email, customer_phone, status, service_price, total_price, discount_amount, tax_amount, tip_amount, currency, payment_status, payment_method, payment_reference, paid_at, notes, special_requests, internal_notes, confirmation_method, confirmation_sent_at, reminder_sent_at, scheduled_start_time, scheduled_end_time, actual_start_time, actual_end_time, cancelled_at, cancelled_by, cancellation_reason, cancellation_fee, booking_source, referral_source, utm_campaign, ml_prediction_score, customer_segment, booking_value_score, created_at, updated_at) VALUES

-- Completed Bookings (for review generation)
('booking-uuid-001', 'BK2024001', 2, 1, 1, 'Executive Business Cut', 'Haircuts', 45, 'John Doe', 'john.doe@email.com', '+1-555-0102', 'completed', 45.00, 54.00, 5.00, 4.00, 10.00, 'USD', 'paid', 'credit_card', 'ch_1234567890', NOW() - INTERVAL '5 days', 'First time customer, looking for professional look', 'Please use medium length on top', 'Customer very satisfied, will return', 'email', NOW() - INTERVAL '7 days', NOW() - INTERVAL '1 day', NOW() - INTERVAL '5 days 10:00:00', NOW() - INTERVAL '5 days 10:45:00', NOW() - INTERVAL '5 days 10:02:00', NOW() - INTERVAL '5 days 10:43:00', NULL, NULL, NULL, 0, 'web_app', 'google', 'summer_promo', 0.15, 'new_customer', 85.2, NOW() - INTERVAL '7 days', NOW() - INTERVAL '5 days'),

('booking-uuid-002', 'BK2024002', 3, 2, 11, 'Trendy Fade Design', 'Haircuts', 60, 'Sarah Wilson', 'sarah.wilson@email.com', '+1-555-0103', 'completed', 50.00, 60.00, 5.00, 5.00, 10.00, 'USD', 'paid', 'debit_card', 'ch_1234567891', NOW() - INTERVAL '3 days', 'Looking for something trendy and modern', 'Would like side part design', 'Excellent work, customer thrilled with result', 'sms', NOW() - INTERVAL '5 days', NOW() - INTERVAL '1 day', NOW() - INTERVAL '3 days 14:00:00', NOW() - INTERVAL '3 days 15:00:00', NOW() - INTERVAL '3 days 14:05:00', NOW() - INTERVAL '3 days 14:58:00', NULL, NULL, NULL, 0, 'mobile_app', 'instagram', 'fade_special', 0.22, 'returning_customer', 92.7, NOW() - INTERVAL '5 days', NOW() - INTERVAL '3 days'),

('booking-uuid-003', 'BK2024003', 4, 3, 15, 'Precision Business Cut', 'Haircuts', 40, 'Mike Johnson', 'mike.johnson@email.com', '+1-555-0104', 'completed', 40.00, 50.00, 0.00, 4.00, 6.00, 'USD', 'paid', 'cash', NULL, NOW() - INTERVAL '2 days', 'Need clean professional cut for job interview', '', 'Perfect cut, customer got the job!', 'email', NOW() - INTERVAL '4 days', NOW() - INTERVAL '1 day', NOW() - INTERVAL '2 days 11:00:00', NOW() - INTERVAL '2 days 11:40:00', NOW() - INTERVAL '2 days 11:00:00', NOW() - INTERVAL '2 days 11:38:00', NULL, NULL, NULL, 0, 'web_app', 'google', NULL, 0.08, 'high_value', 88.9, NOW() - INTERVAL '4 days', NOW() - INTERVAL '2 days'),

-- Upcoming Bookings
('booking-uuid-004', 'BK2024004', 2, 1, 1, 'Executive Business Cut', 'Haircuts', 45, 'John Doe', 'john.doe@email.com', '+1-555-0102', 'confirmed', 40.00, 48.00, 5.00, 3.00, 0.00, 'USD', 'paid', 'credit_card', 'ch_1234567892', NOW() - INTERVAL '1 hour', 'Regular maintenance cut', '', 'Returning customer, preferred stylist', 'email', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 hours', NOW() + INTERVAL '1 day 9:00:00', NOW() + INTERVAL '1 day 9:45:00', NULL, NULL, NULL, NULL, NULL, 0, 'web_app', 'returning_customer', NULL, 0.05, 'loyal_customer', 95.4, NOW() - INTERVAL '2 days', NOW()),

('booking-uuid-005', 'BK2024005', 3, 2, 10, 'Trendy Fade Design', 'Haircuts', 60, 'Sarah Wilson', 'sarah.wilson@email.com', '+1-555-0103', 'confirmed', 45.00, 54.00, 5.00, 4.00, 0.00, 'USD', 'paid', 'apple_pay', 'ap_1234567890', NOW() - INTERVAL '30 minutes', 'Touch up from last visit', 'Same style as before', '', 'app', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '1 day 10:00:00', NOW() + INTERVAL '1 day 11:00:00', NULL, NULL, NULL, NULL, NULL, 0, 'mobile_app', 'returning_customer', NULL, 0.03, 'loyal_customer', 97.1, NOW() - INTERVAL '1 day', NOW()),

-- Pending Bookings
('booking-uuid-006', 'BK2024006', NULL, 3, 17, 'Modern Fade Mastery', 'Haircuts', 55, 'Robert Smith', 'robert.smith@email.com', '+1-555-0199', 'pending', 55.00, 66.00, 0.00, 5.50, 0.00, 'USD', 'pending', NULL, NULL, NULL, 'First time booking, heard great reviews', 'Looking for Asian hair specialist', 'Guest booking, needs confirmation', 'email', NOW() - INTERVAL '2 hours', NULL, NOW() + INTERVAL '2 days 11:00:00', NOW() + INTERVAL '2 days 11:55:00', NULL, NULL, NULL, NULL, NULL, 0, 'web_app', 'google', 'asian_specialist', 0.12, 'new_customer', 82.3, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours');

-- =============================================================================
-- 8. REVIEWS SEED DATA
-- =============================================================================

INSERT INTO reviews (booking_id, customer_id, barber_id, overall_rating, service_quality_rating, punctuality_rating, cleanliness_rating, value_for_money_rating, professionalism_rating, title, comment, pros, cons, would_recommend, would_book_again, service_as_expected, duration_accurate, images, is_verified, is_published, moderation_status, moderation_notes, moderated_by, moderated_at, helpful_votes, total_votes, barber_response, barber_response_at, created_at, updated_at) VALUES

-- Reviews for Tony's Barbershop
(1, 2, 1, 5, 5, 5, 5, 4, 5, 'Excellent Traditional Service', 'Tony provided an outstanding haircut with incredible attention to detail. The hot towel treatment was relaxing and the final result exceeded my expectations. Will definitely be back!', 'Professional service, great atmosphere, traditional techniques', 'Slightly pricey but worth it', true, true, true, true, '[]', true, true, 'approved', 'High quality review', 1, NOW() - INTERVAL '4 days', 12, 15, 'Thank you John! It was a pleasure serving you. Looking forward to your next visit!', NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),

(1, 2, 1, 4, 4, 5, 5, 4, 4, 'Great Cut, Professional Service', 'Really happy with my haircut. Tony knows his craft and the barbershop has an authentic feel. Definitely recommend for anyone looking for a classic cut.', 'Skilled barber, clean shop, good value', 'Wait time was a bit long', true, true, true, true, '[]', true, true, 'approved', NULL, 1, NOW() - INTERVAL '4 days', 8, 10, 'Thanks for the feedback! We''re working on reducing wait times.', NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),

-- Reviews for Maria's Salon
(2, 3, 2, 5, 5, 4, 5, 5, 5, 'Amazing Fade Design!', 'Maria is incredibly talented! She understood exactly what I wanted and created a beautiful fade with design elements. The whole experience was fantastic and I love my new look!', 'Creative styling, bilingual service, modern techniques', '', true, true, true, true, '["https://review-photo-1.jpg"]', true, true, 'approved', 'Excellent detailed review', 1, NOW() - INTERVAL '2 days', 18, 20, '¡Muchas gracias Sarah! Me encanta trabajar con clientes que valoran la creatividad.', NOW() - INTERVAL '1 day', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),

-- Reviews for David's Precision Cuts
(3, 4, 3, 5, 5, 5, 5, 5, 5, 'Perfect Business Cut', 'David gave me exactly what I needed for my job interview. The precision and attention to detail were remarkable. Got the job too! Highly recommend his services.', 'Precise cutting, professional result, great consultation', '', true, true, true, true, '[]', true, true, 'approved', 'Success story review', 1, NOW() - INTERVAL '1 day', 22, 24, 'Congratulations on getting the job Mike! So happy I could help you look your best.', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),

-- Additional historical reviews for rating buildup
(1, 2, 1, 4, 4, 4, 5, 4, 4, 'Good Service', 'Solid haircut, will return', 'Professional', 'None really', true, true, true, true, '[]', true, true, 'approved', NULL, 1, NOW() - INTERVAL '10 days', 5, 6, NULL, NULL, NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),

(2, 3, 2, 4, 4, 3, 4, 4, 4, 'Creative Stylist', 'Maria has great ideas for modern styles', 'Modern approach', 'Took longer than expected', true, true, true, false, '[]', true, true, 'approved', NULL, 1, NOW() - INTERVAL '8 days', 3, 4, NULL, NULL, NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),

(3, 4, 3, 5, 5, 5, 5, 4, 5, 'Outstanding Precision', 'David is a master of his craft', 'Perfect technique', '', true, true, true, true, '[]', true, true, 'approved', NULL, 1, NOW() - INTERVAL '6 days', 15, 16, 'Thank you for the kind words!', NOW() - INTERVAL '5 days', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days');

-- =============================================================================
-- 9. BARBER AVAILABILITY SEED DATA
-- =============================================================================

INSERT INTO barber_availability (barber_id, date, day_of_week, start_time, end_time, availability_type, is_recurring, recurring_pattern, recurring_end_date, notes, blocked_reason, created_at, updated_at) VALUES

-- Tony's Regular Schedule (recurring weekly)
(1, CURRENT_DATE, EXTRACT(DOW FROM CURRENT_DATE), TIME '09:00:00', TIME '18:00:00', 'available', true, 'weekly', CURRENT_DATE + INTERVAL '365 days', 'Regular business hours', NULL, NOW(), NOW()),
(1, CURRENT_DATE + INTERVAL '1 day', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '1 day'), TIME '09:00:00', TIME '18:00:00', 'available', true, 'weekly', CURRENT_DATE + INTERVAL '365 days', 'Regular business hours', NULL, NOW(), NOW()),

-- Maria's Schedule (closed Mondays)
(2, CURRENT_DATE + INTERVAL '2 days', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '2 days'), TIME '10:00:00', TIME '19:00:00', 'available', true, 'weekly', CURRENT_DATE + INTERVAL '365 days', 'Regular business hours', NULL, NOW(), NOW()),
(2, CURRENT_DATE + INTERVAL '3 days', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '3 days'), TIME '10:00:00', TIME '19:00:00', 'available', true, 'weekly', CURRENT_DATE + INTERVAL '365 days', 'Regular business hours', NULL, NOW(), NOW()),

-- David's Schedule
(3, CURRENT_DATE, EXTRACT(DOW FROM CURRENT_DATE), TIME '10:00:00', TIME '18:00:00', 'available', true, 'weekly', CURRENT_DATE + INTERVAL '365 days', 'Regular business hours', NULL, NOW(), NOW()),
(3, CURRENT_DATE + INTERVAL '1 day', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '1 day'), TIME '10:00:00', TIME '18:00:00', 'available', true, 'weekly', CURRENT_DATE + INTERVAL '365 days', 'Regular business hours', NULL, NOW(), NOW()),

-- Some blocked time (lunch breaks, etc.)
(1, CURRENT_DATE + INTERVAL '1 day', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '1 day'), TIME '12:00:00', TIME '13:00:00', 'break', false, 'none', NULL, 'Lunch break', NULL, NOW(), NOW()),
(2, CURRENT_DATE + INTERVAL '2 days', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '2 days'), TIME '13:00:00', TIME '14:00:00', 'break', false, 'none', NULL, 'Lunch break', NULL, NOW(), NOW()),
(3, CURRENT_DATE + INTERVAL '1 day', EXTRACT(DOW FROM CURRENT_DATE + INTERVAL '1 day'), TIME '12:30:00', TIME '13:30:00', 'break', false, 'none', NULL, 'Lunch break', NULL, NOW(), NOW());

-- =============================================================================
-- 10. NOTIFICATIONS SEED DATA
-- =============================================================================

INSERT INTO notifications (user_id, title, message, type, channels, status, sent_at, delivered_at, read_at, related_entity_type, related_entity_id, data, priority, scheduled_for, expires_at, created_at) VALUES

-- Booking confirmations
(2, 'Booking Confirmed', 'Your appointment with Tony''s Classic Barbershop is confirmed for tomorrow at 9:00 AM', 'booking_confirmation', '["email", "push"]', 'delivered', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day', 'booking', 4, '{"booking_number": "BK2024004", "barber_name": "Tony Soprano", "service": "Executive Business Cut"}', 'normal', NOW() - INTERVAL '2 days', NOW() + INTERVAL '7 days', NOW() - INTERVAL '2 days'),

(3, 'Booking Confirmed', 'Your fade appointment with Maria is set for tomorrow at 10:00 AM', 'booking_confirmation', '["email", "sms", "push"]', 'delivered', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NULL, 'booking', 5, '{"booking_number": "BK2024005", "barber_name": "Maria Gonzalez", "service": "Trendy Fade Design"}', 'normal', NOW() - INTERVAL '1 day', NOW() + INTERVAL '7 days', NOW() - INTERVAL '1 day'),

-- Reminders
(2, 'Appointment Reminder', 'Reminder: You have an appointment tomorrow at 9:00 AM with Tony''s Classic Barbershop', 'booking_reminder', '["email", "push"]', 'delivered', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour', 'booking', 4, '{"booking_number": "BK2024004", "reminder_type": "24_hour"}', 'normal', NOW() - INTERVAL '2 hours', NOW() + INTERVAL '2 days', NOW() - INTERVAL '2 hours'),

(3, 'Appointment Reminder', 'Don''t forget your fade appointment tomorrow at 10:00 AM with Maria', 'booking_reminder', '["sms", "push"]', 'delivered', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NULL, 'booking', 5, '{"booking_number": "BK2024005", "reminder_type": "24_hour"}', 'normal', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '2 days', NOW() - INTERVAL '1 hour'),

-- Review requests
(2, 'How was your experience?', 'We hope you loved your haircut! Please take a moment to leave a review about your experience with Tony.', 'review_request', '["email"]', 'delivered', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days', 'booking', 1, '{"booking_number": "BK2024001", "review_link": "https://app.barbershop.com/reviews/new?booking=1"}', 'low', NOW() - INTERVAL '4 days', NOW() + INTERVAL '14 days', NOW() - INTERVAL '4 days'),

(3, 'Share your experience', 'How did you like your new fade? Your feedback helps other customers and supports Maria''s business.', 'review_request', '["email", "push"]', 'delivered', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', 'booking', 2, '{"booking_number": "BK2024002", "review_link": "https://app.barbershop.com/reviews/new?booking=2"}', 'low', NOW() - INTERVAL '2 days', NOW() + INTERVAL '14 days', NOW() - INTERVAL '2 days'),

-- Promotional notifications
(2, 'Special Offer Just for You!', 'Tony is offering $5 off executive cuts this week. Book your next appointment and save!', 'promotion', '["email"]', 'delivered', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days', NULL, NULL, NULL, '{"discount_amount": 5, "promo_code": "EXEC5OFF", "valid_until": "2024-12-31"}', 'low', NOW() - INTERVAL '3 days', NOW() + INTERVAL '7 days', NOW() - INTERVAL '3 days'),

(3, 'New Client Special', 'Welcome back! Maria is offering 10% off fade cuts for returning customers this month.', 'promotion', '["email", "push"]', 'delivered', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days', NULL, NULL, '{"discount_percent": 10, "promo_code": "FADE10", "valid_until": "2024-12-31"}', 'low', NOW() - INTERVAL '5 days', NOW() + INTERVAL '30 days', NOW() - INTERVAL '5 days');

-- =============================================================================
-- SUMMARY AND VERIFICATION QUERIES
-- =============================================================================

-- Check data integrity and counts
/*
-- Verify user counts by type
SELECT user_type, COUNT(*) as count, 
       COUNT(CASE WHEN status = 'active' THEN 1 END) as active_count
FROM users 
GROUP BY user_type;

-- Verify barber setup
SELECT b.shop_name, u.name as owner_name, b.status, b.is_verified,
       (SELECT COUNT(*) FROM barber_services bs WHERE bs.barber_id = b.id AND bs.is_active = true) as active_services,
       b.rating, b.total_reviews, b.total_bookings
FROM barbers b
JOIN users u ON b.user_id = u.id;

-- Verify service catalog
SELECT sc.name as category, COUNT(s.id) as service_count,
       COUNT(CASE WHEN s.is_active AND s.is_approved THEN 1 END) as active_count
FROM service_categories sc
LEFT JOIN services s ON sc.id = s.category_id
GROUP BY sc.id, sc.name
ORDER BY sc.sort_order;

-- Verify booking statuses
SELECT status, COUNT(*) as count, 
       ROUND(AVG(total_price), 2) as avg_price
FROM bookings 
GROUP BY status
ORDER BY 
    CASE status 
        WHEN 'completed' THEN 1 
        WHEN 'confirmed' THEN 2 
        WHEN 'pending' THEN 3 
        ELSE 4 
    END;

-- Verify review ratings
SELECT barber_id, 
       (SELECT shop_name FROM barbers WHERE id = barber_id) as shop_name,
       COUNT(*) as review_count,
       ROUND(AVG(overall_rating::numeric), 2) as avg_rating,
       COUNT(CASE WHEN overall_rating >= 4 THEN 1 END) as positive_reviews
FROM reviews 
WHERE is_published = true
GROUP BY barber_id
ORDER BY avg_rating DESC;

-- Check upcoming appointments
SELECT b.booking_number, 
       u.name as customer_name,
       br.shop_name,
       bk.service_name,
       bk.scheduled_start_time,
       bk.status
FROM bookings bk
JOIN users u ON bk.customer_id = u.id
JOIN barbers br ON bk.barber_id = br.id
WHERE bk.scheduled_start_time > NOW()
  AND bk.status IN ('confirmed', 'pending')
ORDER BY bk.scheduled_start_time;
*/

-- =============================================================================
-- UPDATE BARBER STATISTICS BASED ON SEED DATA
-- =============================================================================

-- Update Tony's statistics
UPDATE barbers SET 
    rating = 4.7,
    total_reviews = 4,
    total_bookings = 234 + 156 + 89,
    response_time_minutes = 15,
    acceptance_rate = 94.5,
    cancellation_rate = 2.8
WHERE id = 1;

-- Update Maria's statistics  
UPDATE barbers SET
    rating = 4.5,
    total_reviews = 2,
    total_bookings = 187 + 92,
    response_time_minutes = 25,
    acceptance_rate = 89.2,
    cancellation_rate = 5.1
WHERE id = 2;

-- Update David's statistics
UPDATE barbers SET
    rating = 4.9,
    total_reviews = 2,
    total_bookings = 298 + 267 + 145,
    response_time_minutes = 12,
    acceptance_rate = 96.8,
    cancellation_rate = 2.1
WHERE id = 3;

-- Update service category statistics
UPDATE service_categories SET
    service_count = (SELECT COUNT(*) FROM services WHERE category_id = service_categories.id AND is_active = true),
    barber_count = (SELECT COUNT(DISTINCT bs.barber_id) 
                   FROM barber_services bs 
                   JOIN services s ON bs.service_id = s.id 
                   WHERE s.category_id = service_categories.id AND bs.is_active = true),
    average_price = (SELECT COALESCE(AVG(bs.price), 0) 
                    FROM barber_services bs 
                    JOIN services s ON bs.service_id = s.id 
                    WHERE s.category_id = service_categories.id AND bs.is_active = true);

-- Create indexes for better performance (if not already created)
CREATE INDEX IF NOT EXISTS idx_bookings_customer_id ON bookings(customer_id);
CREATE INDEX IF NOT EXISTS idx_bookings_barber_id ON bookings(barber_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_scheduled_start_time ON bookings(scheduled_start_time);
CREATE INDEX IF NOT EXISTS idx_reviews_barber_id ON reviews(barber_id);
CREATE INDEX IF NOT EXISTS idx_reviews_customer_id ON reviews(customer_id);
CREATE INDEX IF NOT EXISTS idx_reviews_is_published ON reviews(is_published);
CREATE INDEX IF NOT EXISTS idx_barber_services_barber_id ON barber_services(barber_id);
CREATE INDEX IF NOT EXISTS idx_barber_services_is_active ON barber_services(is_active);
CREATE INDEX IF NOT EXISTS idx_time_slots_barber_id ON time_slots(barber_id);
CREATE INDEX IF NOT EXISTS idx_time_slots_start_time ON time_slots(start_time);
CREATE INDEX IF NOT EXISTS idx_time_slots_is_available ON time_slots(is_available);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);

-- =============================================================================
-- SEED SCRIPT COMPLETION MESSAGE
-- =============================================================================

DO $
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'DATABASE SEEDING COMPLETED SUCCESSFULLY!';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Seeded Data Summary:';
    RAISE NOTICE '- Users: % (% customers, % barbers, % admins)', 
        (SELECT COUNT(*) FROM users),
        (SELECT COUNT(*) FROM users WHERE user_type = 'customer'),
        (SELECT COUNT(*) FROM users WHERE user_type = 'barber'),
        (SELECT COUNT(*) FROM users WHERE user_type = 'admin');
    RAISE NOTICE '- Barbers: % active barbershops', 
        (SELECT COUNT(*) FROM barbers WHERE status = 'active');
    RAISE NOTICE '- Services: % global services in % categories', 
        (SELECT COUNT(*) FROM services WHERE is_active = true),
        (SELECT COUNT(*) FROM service_categories WHERE is_active = true);
    RAISE NOTICE '- Barber Services: % total offerings', 
        (SELECT COUNT(*) FROM barber_services WHERE is_active = true);
    RAISE NOTICE '- Bookings: % total (% completed, % confirmed, % pending)', 
        (SELECT COUNT(*) FROM bookings),
        (SELECT COUNT(*) FROM bookings WHERE status = 'completed'),
        (SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'),
        (SELECT COUNT(*) FROM bookings WHERE status = 'pending');
    RAISE NOTICE '- Reviews: % published reviews', 
        (SELECT COUNT(*) FROM reviews WHERE is_published = true);
    RAISE NOTICE '- Time Slots: % available slots', 
        (SELECT COUNT(*) FROM time_slots WHERE is_available = true);
    RAISE NOTICE '- Notifications: % total notifications', 
        (SELECT COUNT(*) FROM notifications);
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Test Login Credentials:';
    RAISE NOTICE 'Admin: admin@barbershop.com / password123';
    RAISE NOTICE 'Customer: john.doe@email.com / password123';
    RAISE NOTICE 'Barber: tony.soprano@barbershop.com / password123';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Your barbershop database is ready for testing!';
    RAISE NOTICE '=============================================================================';
END $;-- Database Seed Scripts for Barbershop Application
-- Run these scripts in order to populate your database with sample data

-- =============================================================================
-- 1. USERS SEED DATA
-- =============================================================================

-- Insert sample users (customers, barbers, admins)
INSERT INTO users (uuid, email, password_hash, name, phone, user_type, status, email_verified, phone_verified, date_of_birth, gender, profile_picture_url, address, city, state, country, postal_code, latitude, longitude, preferences, notification_settings, created_at, updated_at, last_login_at) VALUES

-- Admin Users
('550e8400-e29b-41d4-a716-446655440001', 'admin@barbershop.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'System Administrator', '+1-555-0101', 'admin', 'active', true, true, '1985-03-15', 'male', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150', '123 Admin St', 'New York', 'NY', 'USA', '10001', 40.7128, -74.0060, '{"theme": "dark", "language": "en"}', '{"email": true, "sms": true, "push": true}', NOW() - INTERVAL '30 days', NOW(), NOW() - INTERVAL '1 day'),

-- Customer Users
('550e8400-e29b-41d4-a716-446655440002', 'john.doe@email.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'John Doe', '+1-555-0102', 'customer', 'active', true, true, '1990-07-22', 'male', 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=150', '456 Main St', 'New York', 'NY', 'USA', '10002', 40.7589, -73.9851, '{"preferred_barber": 1, "hair_type": "straight"}', '{"email": true, "sms": false, "push": true}', NOW() - INTERVAL '15 days', NOW(), NOW() - INTERVAL '2 hours'),

('550e8400-e29b-41d4-a716-446655440003', 'sarah.wilson@email.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'Sarah Wilson', '+1-555-0103', 'customer', 'active', true, false, '1988-11-08', 'female', 'https://images.unsplash.com/photo-1494790108755-2616b612b6e5?w=150', '789 Oak Ave', 'Brooklyn', 'NY', 'USA', '11201', 40.6892, -73.9442, '{"hair_type": "curly", "preferred_style": "modern"}', '{"email": true, "sms": true, "push": false}', NOW() - INTERVAL '10 days', NOW(), NOW() - INTERVAL '1 hour'),

('550e8400-e29b-41d4-a716-446655440004', 'mike.johnson@email.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'Mike Johnson', '+1-555-0104', 'customer', 'active', true, true, '1995-02-14', 'male', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150', '321 Pine St', 'Manhattan', 'NY', 'USA', '10003', 40.7505, -73.9934, '{"hair_type": "thick", "beard_style": "full"}', '{"email": false, "sms": true, "push": true}', NOW() - INTERVAL '5 days', NOW(), NOW() - INTERVAL '30 minutes'),

-- Barber Users
('550e8400-e29b-41d4-a716-446655440005', 'tony.soprano@barbershop.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'Tony Soprano', '+1-555-0105', 'barber', 'active', true, true, '1980-09-12', 'male', 'https://images.unsplash.com/photo-1521119989659-a83eee488004?w=150', '100 Barber Row', 'New York', 'NY', 'USA', '10001', 40.7128, -74.0060, '{"specialties": ["classic_cuts", "beard_styling"], "languages": ["en", "it"]}', '{"email": true, "sms": true, "push": true}', NOW() - INTERVAL '20 days', NOW(), NOW() - INTERVAL '15 minutes'),

('550e8400-e29b-41d4-a716-446655440006', 'maria.gonzalez@barbershop.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'Maria Gonzalez', '+1-555-0106', 'barber', 'active', true, true, '1987-04-03', 'female', 'https://images.unsplash.com/photo-1580618672591-eb180b1a973f?w=150', '200 Style Ave', 'Brooklyn', 'NY', 'USA', '11201', 40.6892, -73.9442, '{"specialties": ["modern_cuts", "color"], "languages": ["en", "es"]}', '{"email": true, "sms": false, "push": true}', NOW() - INTERVAL '25 days', NOW(), NOW() - INTERVAL '1 hour'),

('550e8400-e29b-41d4-a716-446655440007', 'david.kim@barbershop.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.', 'David Kim', '+1-555-0107', 'barber', 'active', true, true, '1985-12-18', 'male', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150', '300 Trend St', 'Queens', 'NY', 'USA', '11101', 40.7505, -73.9934, '{"specialties": ["asian_styles", "precision_cuts"], "languages": ["en", "ko"]}', '{"email": true, "sms": true, "push": true}', NOW() - INTERVAL '18 days', NOW(), NOW() - INTERVAL '30 minutes');

-- =============================================================================
-- 2. SERVICE CATEGORIES SEED DATA
-- =============================================================================

INSERT INTO service_categories (name, slug, description, parent_category_id, level, category_path, icon_url, color_hex, image_url, sort_order, is_active, is_featured, meta_title, meta_description, keywords, service_count, barber_count, average_price, popularity_score, created_at, updated_at) VALUES

-- Main Categories
('Haircuts', 'haircuts', 'Professional hair cutting services for all styles and preferences', NULL, 1, 'haircuts', 'https://cdn.barbershop.com/icons/haircut.svg', '#2563EB', 'https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=300', 1, true, true, 'Professional Haircuts | Barbershop', 'Expert haircut services including classic, modern, and trendy styles', '["haircut", "styling", "mens", "womens"]', 8, 15, 35.00, 95.5, NOW(), NOW()),

('Beard & Mustache', 'beard-mustache', 'Expert beard trimming, shaping, and mustache styling services', NULL, 1, 'beard-mustache', 'https://cdn.barbershop.com/icons/beard.svg', '#DC2626', 'https://images.unsplash.com/photo-1621605815971-fbc98d665033?w=300', 2, true, true, 'Beard & Mustache Services | Barbershop', 'Professional beard trimming and mustache styling services', '["beard", "mustache", "trimming", "shaping"]', 5, 12, 25.00, 88.2, NOW(), NOW()),

('Hair Styling', 'hair-styling', 'Creative hair styling and special occasion styling services', NULL, 1, 'hair-styling', 'https://cdn.barbershop.com/icons/styling.svg', '#7C3AED', 'https://images.unsplash.com/photo-1562004760-acb5f2f1dfef?w=300', 3, true, true, 'Hair Styling Services | Barbershop', 'Professional hair styling for special occasions and everyday looks', '["styling", "formal", "special", "occasion"]', 4, 8, 45.00, 78.9, NOW(), NOW()),

('Hair Treatments', 'hair-treatments', 'Therapeutic hair and scalp treatments for hair health', NULL, 1, 'hair-treatments', 'https://cdn.barbershop.com/icons/treatment.svg', '#059669', 'https://images.unsplash.com/photo-1560066984-138dadb4c035?w=300', 4, true, false, 'Hair Treatments | Barbershop', 'Professional hair and scalp treatments for optimal hair health', '["treatment", "scalp", "therapy", "health"]', 3, 5, 60.00, 65.4, NOW(), NOW()),

-- Sub Categories
('Classic Cuts', 'classic-cuts', 'Traditional and timeless haircut styles', 1, 2, 'haircuts/classic-cuts', NULL, '#1E40AF', NULL, 1, true, false, NULL, NULL, '["classic", "traditional", "timeless"]', 3, 10, 30.00, 85.2, NOW(), NOW()),

('Modern Cuts', 'modern-cuts', 'Contemporary and trendy haircut styles', 1, 2, 'haircuts/modern-cuts', NULL, '#3B82F6', NULL, 2, true, true, NULL, NULL, '["modern", "trendy", "contemporary"]', 5, 12, 40.00, 92.8, NOW(), NOW());

-- =============================================================================
-- 3. SERVICES SEED DATA (Global Catalog)
-- =============================================================================

INSERT INTO services (uuid, name, slug, short_description, detailed_description, category_id, service_type, complexity, skill_level_required, default_duration_min, default_duration_max, suggested_price_min, suggested_price_max, currency, target_gender, target_age_min, target_age_max, hair_types, requires_consultation, required_tools, required_products, required_certifications, allergen_warnings, health_precautions, requires_health_check, image_url, gallery_images, video_url, tags, search_keywords, meta_description, has_variations, allows_add_ons, global_popularity_score, total_global_bookings, average_global_rating, total_global_reviews, is_active, is_approved, approval_notes, created_at, updated_at, created_by, last_modified_by, version, change_log) VALUES

-- Haircut Services
('service-uuid-001', 'Classic Business Cut', 'classic-business-cut', 'Professional business haircut for the modern gentleman', 'A timeless, professional haircut perfect for business environments. Includes precision cutting, styling, and finishing touches for a polished look that commands respect.', 5, 'haircut', 3, 'intermediate', 30, 45, 25.00, 40.00, 'USD', 'male', 18, 65, '["straight", "wavy", "thick", "all"]', false, '["scissors", "comb", "clippers", "styling_brush"]', '["shampoo", "conditioner", "styling_gel"]', '[]', '[]', '[]', false, 'https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=400', '["https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=400"]', NULL, '["classic", "business", "professional", "conservative"]', '["business cut", "professional haircut", "classic style"]', 'Professional classic business haircut for men', true, true, 92.5, 2847, 4.6, 1205, true, true, 'Popular classic style', NOW() - INTERVAL '60 days', NOW(), 1, 1, 1, '{"initial": "Service created"}'),

('service-uuid-002', 'Modern Fade Cut', 'modern-fade-cut', 'Trendy fade haircut with modern styling', 'Contemporary fade cut featuring gradual length transition from short sides to longer top. Customizable fade height and styling options for a fresh, modern look.', 6, 'haircut', 4, 'advanced', 45, 60, 35.00, 55.00, 'USD', 'male', 16, 45, '["straight", "wavy", "curly", "all"]', true, '["clippers", "scissors", "razor", "comb", "styling_brush"]', '["shampoo", "pomade", "hair_wax"]', '[]', '[]', '[]', false, 'https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=400', '["https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=400"]', 'https://youtube.com/watch?v=fade-tutorial', '["modern", "fade", "trendy", "gradient"]', '["fade cut", "modern fade", "skin fade"]', 'Modern fade haircut with precision styling', true, true, 96.8, 3652, 4.8, 1847, true, true, 'Highly requested modern style', NOW() - INTERVAL '45 days', NOW(), 1, 1, 1, '{"initial": "Service created"}'),

('service-uuid-003', 'Beard Trim & Shape', 'beard-trim-shape', 'Professional beard trimming and shaping service', 'Expert beard trimming and shaping to maintain your desired look. Includes precise trimming, edge work, and styling with premium beard products.', 2, 'grooming', 3, 'intermediate', 20, 30, 15.00, 30.00, 'USD', 'male', 18, 70, '["all"]', false, '["beard_trimmer", "scissors", "razor", "comb"]', '["beard_oil", "beard_balm", "aftershave"]', '[]', '["fragrance"]', '["skin_sensitivity"]', false, 'https://images.unsplash.com/photo-1621605815971-fbc98d665033?w=400', '["https://images.unsplash.com/photo-1621605815971-fbc98d665033?w=400"]', NULL, '["beard", "trim", "shape", "grooming"]', '["beard trim", "beard shaping", "facial hair"]', 'Professional beard trimming and shaping service', true, true, 89.2, 1956, 4.5, 876, true, true, 'Essential grooming service', NOW() - INTERVAL '30 days', NOW(), 1, 1, 1, '{"initial": "Service created"}'),

('service-uuid-004', 'Hot Towel Shave', 'hot-towel-shave', 'Traditional hot towel straight razor shave', 'Classic barbershop experience with hot towel treatment, premium shaving cream, and straight razor shave. Includes aftershave treatment and moisturizing.', 2, 'grooming', 5, 'expert', 45, 60, 40.00, 70.00, 'USD', 'male', 18, 80, '["all"]', true, '["straight_razor", "hot_towels", "shaving_brush", "strop"]', '["shaving_cream", "pre_shave_oil", "aftershave", "moisturizer"]', '["barbering_license"]', '["fragrance", "lanolin"]', '["skin_sensitivity", "blood_thinners"]', true, 'https://images.unsplash.com/photo-1585747860715-2ba37e788b70?w=400', '["https://images.unsplash.com/photo-1585747860715-2ba37e788b70?w=400"]', NULL, '["traditional", "shave", "hot_towel", "luxury"]', '["hot towel shave", "straight razor", "traditional shave"]', 'Traditional hot towel straight razor shave experience', false, true, 78.4, 892, 4.9, 425, true, true, 'Premium traditional service', NOW() - INTERVAL '20 days', NOW(), 1, 1, 1, '{"initial": "Service created"}'),

('service-uuid-005', 'Kids Haircut', 'kids-haircut', 'Fun and gentle haircuts for children', 'Child-friendly haircut service with patient approach and fun atmosphere. Includes basic styling and can accommodate fidgety children with toys and entertainment.', 1, 'haircut', 2, 'beginner', 20, 30, 15.00, 25.00, 'USD', 'all', 3, 17, '["straight", "wavy", "curly", "all"]', false, '["scissors", "clippers", "comb", "cape"]', '["mild_shampoo", "detangling_spray"]', '[]', '[]', '[]', false, 'https://images.unsplash.com/photo-1564463489817-3f6eccb47b89?w=400', '["https://images.unsplash.com/photo-1564463489817-3f6eccb47b89?w=400"]', NULL, '["kids", "children", "gentle", "fun"]', '["kids haircut", "children haircut", "child friendly"]', 'Fun and gentle haircuts for children of all ages', true, false, 85.7, 1653, 4.4, 743, true, true, 'Family-friendly service', NOW() - INTERVAL '40 days', NOW(), 1, 1, 1, '{"initial": "Service created"}'),

('service-uuid-006', 'Hair Wash & Style', 'hair-wash-style', 'Complete hair washing and styling service', 'Thorough hair washing with premium products followed by professional styling. Perfect for special occasions or regular maintenance.', 3, 'styling', 2, 'intermediate', 30, 45, 20.00, 35.00, 'USD', 'all', 12, 80, '["straight", "wavy", "curly", "coily", "all"]', false, '["shampoo_bowl", "hair_dryer", "styling_brush", "curling_iron"]', '["professional_shampoo", "conditioner", "styling_mousse", "hairspray"]', '[]', '["sulfates", "parabens"]', '["scalp_sensitivity"]', false, 'https://images.unsplash.com/photo-1562004760-acb5f2f1dfef?w=400', '["https://images.unsplash.com/photo-1562004760-acb5f2f1dfef?w=400"]', NULL, '["wash", "style", "maintenance", "care"]', '["hair wash", "styling", "hair care"]', 'Professional hair washing and styling service', true, true, 82.3, 1247, 4.3, 567, true, true, 'Regular maintenance service', NOW() - INTERVAL '35 days', NOW(), 1, 1, 1, '{"initial": "Service created"}');

-- =============================================================================
-- 4. BARBERS SEED DATA
-- =============================================================================

INSERT INTO barbers (user_id, uuid, shop_name, business_name, business_registration_number, tax_id, address, address_line_2, city, state, country, postal_code, latitude, longitude, phone, business_email, website_url, description, years_experience, specialties, certifications, languages_spoken, profile_image_url, cover_image_url, gallery_images, working_hours, rating, total_reviews, total_bookings, response_time_minutes, acceptance_rate, cancellation_rate, status, is_verified, verification_date, verification_notes, advance_booking_days, min_booking_notice_hours, auto_accept_bookings, instant_booking_enabled, commission_rate, payout_method, payout_details, created_at, updated_at, last_active_at) VALUES

-- Tony's Classic Barbershop
(5, '550e8400-e29b-41d4-a716-446655440101', 'Tony''s Classic Barbershop', 'Antonio Soprano Barber Services LLC', 'LLC2023001', 'TAX123456789', '100 Barber Row', 'Suite 101', 'New York', 'NY', 'USA', '10001', 40.7128, -74.0060, '+1-555-0105', 'tony@classicbarbershop.com', 'https://tonysclassicbarbershop.com', 'Traditional Italian barbershop offering classic cuts, hot towel shaves, and authentic grooming services. Family-owned business with 15+ years of experience serving NYC.', 15, '["classic_cuts", "hot_towel_shave", "beard_styling", "traditional_grooming"]', '["Master_Barber_Certificate", "Traditional_Shaving_Certification"]', '["English", "Italian"]', 'https://images.unsplash.com/photo-1521119989659-a83eee488004?w=300', 'https://images.unsplash.com/photo-1585747860715-2ba37e788b70?w=800', '["https://images.unsplash.com/photo-1585747860715-2ba37e788b70?w=400", "https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=400", "https://images.unsplash.com/photo-1621605815971-fbc98d665033?w=400"]', '{"monday": {"open": "09:00", "close": "18:00"}, "tuesday": {"open": "09:00", "close": "18:00"}, "wednesday": {"open": "09:00", "close": "18:00"}, "thursday": {"open": "09:00", "close": "19:00"}, "friday": {"open": "09:00", "close": "19:00"}, "saturday": {"open": "08:00", "close": "17:00"}, "sunday": {"closed": true}}', 4.8, 127, 892, 15, 94.5, 3.2, 'active', true, NOW() - INTERVAL '25 days', 'Excellent traditional barber with strong customer base', 30, 2, false, true, 15.0, 'bank_transfer', '{"bank": "Chase Bank", "account_type": "business_checking"}', NOW() - INTERVAL '20 days', NOW(), NOW() - INTERVAL '15 minutes'),

-- Maria's Modern Salon
(6, '550e8400-e29b-41d4-a716-446655440102', 'Maria''s Modern Salon', 'Maria Gonzalez Hair Studio Inc', 'INC2023002', 'TAX987654321', '200 Style Ave', 'Floor 2', 'Brooklyn', 'NY', 'USA', '11201', 40.6892, -73.9442, '+1-555-0106', 'maria@modernstyles.com', 'https://mariasmodernsalon.com', 'Contemporary hair salon specializing in modern cuts, color treatments, and creative styling. Bilingual service with focus on latest trends and techniques.', 8, '["modern_cuts", "color_treatments", "creative_styling", "womens_cuts"]', '["Cosmetology_License", "Color_Specialist_Certification", "Balayage_Expert"]', '["English", "Spanish"]', 'https://images.unsplash.com/photo-1580618672591-eb180b1a973f?w=300', 'https://images.unsplash.com/photo-1562004760-acb5f2f1dfef?w=800', '["https://images.unsplash.com/photo-1562004760-acb5f2f1dfef?w=400", "https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=400", "https://images.unsplash.com/photo-1560066984-138dadb4c035?w=400"]', '{"monday": {"closed": true}, "tuesday": {"open": "10:00", "close": "19:00"}, "wednesday": {"open": "10:00", "close": "19:00"}, "thursday": {"open": "10:00", "close": "20:00"}, "friday": {"open": "10:00", "close": "20:00"}, "saturday": {"open": "09:00", "close": "18:00"}, "sunday": {"open": "11:00", "close": "16:00"}}', 4.6, 89, 654, 25, 89.2, 5.1, 'active', true, NOW() - INTERVAL '30 days', 'Talented stylist with modern approach', 45, 4, true, false, 18.0, 'paypal', '{"email": "maria.payments@email.com"}', NOW() - INTERVAL '25 days', NOW(), NOW() - INTERVAL '1 hour'),

-- David's Precision Cuts
(7, '550e8400-e29b-41d4-a716-446655440103', 'David''s Precision Cuts', 'DK Hair Design Studio', 'LLC2023003', 'TAX456789123', '300 Trend St', NULL, 'Queens', 'NY', 'USA', '11101', 40.7505, -73.9934, '+1-555-0107', 'david@precisioncuts.com', 'https://davidsprecisioncuts.com', 'Precision cutting specialist focusing on Asian hair types and modern styling techniques. Attention to detail and personalized service guaranteed.', 10, '["precision_cuts", "asian_hair_specialist", "modern_styling", "fade_expert"]', '["Advanced_Cutting_Certification", "Asian_Hair_Specialist", "Fade_Master"]', '["English", "Korean", "Mandarin"]', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=300', 'https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=800', '["https://images.unsplash.com/photo-1503951914875-452162b0f3f1?w=400", "https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=400"]', '{"monday": {"open": "10:00", "close": "18:00"}, "tuesday": {"open": "10:00", "close": "18:00"}, "wednesday": {"open": "10:00", "close": "18:00"}, "thursday": {"open": "10:00", "close": "19:00"}, "friday": {"open": "10:00", "close": "19:00"}, "saturday": {"open": "09:00", "close": "17:00"}, "sunday": {"closed": true}}', 4.9, 156, 1123, 12, 96.8, 2.1, 'active', true, NOW() - INTERVAL '18 days', 'Exceptional precision and technique', 60, 3, false, true, 20.0, 'stripe', '{"account_id": "acct_1234567890"}', NOW() - INTERVAL '18 days', NOW(), NOW() - INTERVAL '30 minutes');

-- =============================================================================
-- 5. BARBER SERVICES SEED DATA
-- =============================================================================

INSERT INTO barber_services (barber_id, service_id, custom_name, custom_description, price, max_price, currency, discount_price, discount_valid_until, estimated_duration_min, estimated_duration_max, buffer_time_minutes, advance_notice_hours, max_advance_booking_days, available_days, available_time_slots, requires_consultation, consultation_duration, pre_service_instructions, post_service_care, min_customer_age, max_customer_age, is_seasonal, seasonal_start_month, seasonal_end_month, portfolio_images, before_after_images, total_bookings, total_revenue, average_rating, total_reviews, cancellation_rate, customer_satisfaction, repeat_customer_rate, bookings_last_30_days, revenue_last_30_days, popularity_score, demand_level, is_promotional, promotional_text, promotion_start_date, promotion_end_date, is_featured, display_order, service_note, is_active, paused_reason, paused_until, created_at, updated_at) VALUES

-- Tony's Services
(1, 1, 'Executive Business Cut', 'Premium business haircut with traditional Italian styling techniques', 45.00, 55.00, 'USD', 40.00, NOW() + INTERVAL '7 days', 35, 45, 15, 2, 30, '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]', '{}', false, NULL, 'Please arrive with clean, dry hair', 'Avoid washing hair for 24 hours for best styling results', 16, 65, false, NULL, NULL, '["https://images.unsplash.com/photo-1622286346003-c3748d7d2c34?w=400"]', '["https://before-after-1.jpg", "https://before-after-2.jpg"]', 234, 10530.00, 4.8, 89, 2.1, 96.5, 78.2, 18, 810.00, 94.5, 0.8, true, 'Limited time: $5 off executive cuts!', NOW(), NOW() + INTERVAL '7 days', true, 1, 'Signature service with complimentary hot towel', true, NULL, NULL, NOW() - INTERVAL '20 days', NOW()),

(1, 3, 'Traditional Beard Sculpting', 'Artisanal beard trimming using traditional Italian techniques', 30.00, 40.00, 'USD', NULL, NULL, 25, 35, 10, 2, 30, '["monday", "tuesday", "wednesday